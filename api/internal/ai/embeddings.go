package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
)

// EmbeddingModel names an OpenAI embedding model. Typed string so
// callers don't import openai-go directly. Embeddings use a separate
// SDK service (sdk.Embeddings) than chat completions, so we don't
// alias the chat Model type.
type EmbeddingModel string

const (
	// EmbeddingModelTextEmbedding3Small is the production retrieval
	// embedder (ASK-219 schema is 1536-dim to match). $0.02 / 1M
	// tokens; the cost-log line writes USD at this rate so spend can
	// be ledger-summed without re-deriving the price per row.
	EmbeddingModelTextEmbedding3Small EmbeddingModel = "text-embedding-3-small"

	// EmbeddingModelDefault is the model new callers should use when
	// they don't have a reason to override.
	EmbeddingModelDefault = EmbeddingModelTextEmbedding3Small

	// embedBatchSize caps inputs per OpenAI request. OpenAI accepts
	// up to 2048 inputs in one call but returns the whole response
	// only after every input has embedded -- we keep batches small
	// so a single dropped item doesn't stall a large file. 100 is
	// the ASK-221 ticket's choice; revisit if rate limits push back.
	embedBatchSize = 100

	// embedMaxAttempts caps retry attempts per batch. OpenAI's
	// transient 429 / 5xx are usually cleared within ~5s; we add
	// exponential backoff up to attempt #3 then surface the error.
	embedMaxAttempts = 3

	// embedRetryBaseDelay is the first retry's wait. Doubles each
	// attempt (capped at embedRetryMaxDelay).
	embedRetryBaseDelay = 1 * time.Second
	embedRetryMaxDelay  = 30 * time.Second

	// recordUsageTimeout bounds the ai_usage write so a slow Postgres
	// can't stall the embedding worker indefinitely. Mirrors the 5s
	// envelope ASK-214's chat path uses.
	recordUsageTimeout = 5 * time.Second
)

// embeddingModelPriceUSDPerM maps known embedding model identifiers
// to their per-million-token list price. Adding a new model means
// adding it here so the cost log doesn't silently misreport. An
// unmapped model returns 0 (the cost log will show 0.0 USD with a
// warning) rather than guessing -- prefer "I don't know the price"
// over "I made one up".
var embeddingModelPriceUSDPerM = map[EmbeddingModel]float64{
	EmbeddingModelTextEmbedding3Small: 0.02,
}

// usdForEmbedding returns (cost_usd, knownPrice) for tokens consumed
// by the given model.
func usdForEmbedding(model EmbeddingModel, tokens int64) (float64, bool) {
	rate, ok := embeddingModelPriceUSDPerM[model]
	if !ok {
		return 0, false
	}
	return float64(tokens) * rate / 1_000_000, true
}

// EmbedRequest is the input to Client.Embed.
type EmbedRequest struct {
	// UserID attributes cost + usage to the file owner. Required.
	UserID uuid.UUID
	// Feature tags the calling surface in the cost log. Defaults to
	// FeatureEmbedding when zero.
	Feature Feature
	// Model selects the embedder. Zero value = EmbeddingModelDefault.
	Model EmbeddingModel
	// Inputs is the list of texts to embed. Order is preserved in
	// the returned Vectors slice. Empty Inputs is a no-op success
	// (zero vectors, zero tokens).
	Inputs []string
}

// EmbedResponse carries the embedding vectors plus aggregated usage.
// Vectors[i] corresponds to req.Inputs[i].
type EmbedResponse struct {
	Vectors [][]float32
	Usage   Usage // InputTokens populated; Output/Cache* are zero
}

// Embed batches req.Inputs into OpenAI Embeddings API calls (up to
// embedBatchSize per call), retries 429 / 5xx with exponential
// backoff (capped at embedMaxAttempts), aggregates the resulting
// vectors + token usage, and emits a single cost-log line + an
// ai_usage row tagged FeatureEmbedding. Order is preserved across
// batches.
//
// Errors:
//   - validation (missing UserID, empty input string) → returned
//     immediately, no API call made.
//   - upstream non-retryable → returned with the API error wrapped.
//   - upstream retry budget exhausted → returns the last error.
//
// Cancellation: ctx cancellation propagates to the HTTP client;
// partial usage from completed batches still gets logged + recorded.
func (c *Client) Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error) {
	if req.UserID == uuid.Nil {
		return EmbedResponse{}, errors.New("ai.Embed: UserID required")
	}
	for i, in := range req.Inputs {
		if in == "" {
			return EmbedResponse{}, fmt.Errorf("ai.Embed: input[%d] is empty", i)
		}
	}
	if len(req.Inputs) == 0 {
		return EmbedResponse{}, nil
	}
	feature := req.Feature
	if feature == "" {
		feature = FeatureEmbedding
	}
	model := req.Model
	if model == "" {
		model = EmbeddingModelDefault
	}

	requestID := uuid.NewString()
	out := make([][]float32, 0, len(req.Inputs))
	totalTokens := int64(0)

	for start := 0; start < len(req.Inputs); start += embedBatchSize {
		end := start + embedBatchSize
		if end > len(req.Inputs) {
			end = len(req.Inputs)
		}
		batch := req.Inputs[start:end]

		vecs, tokens, err := c.embedBatchWithRetry(ctx, batch, model)
		if err != nil {
			c.logEmbedCost(req.UserID, feature, model, requestID, totalTokens, err)
			c.recordEmbedUsage(req.UserID, feature, model, requestID, totalTokens)
			return EmbedResponse{}, fmt.Errorf("ai.Embed: batch %d-%d: %w", start, end, err)
		}
		out = append(out, vecs...)
		totalTokens += tokens
	}

	c.logEmbedCost(req.UserID, feature, model, requestID, totalTokens, nil)
	c.recordEmbedUsage(req.UserID, feature, model, requestID, totalTokens)

	return EmbedResponse{
		Vectors: out,
		Usage:   Usage{InputTokens: totalTokens},
	}, nil
}

// embedBatchWithRetry sends one batch and retries 429 / 5xx with
// exponential backoff. Non-retryable errors (4xx other than 429,
// validation) bail immediately.
func (c *Client) embedBatchWithRetry(ctx context.Context, batch []string, model EmbeddingModel) ([][]float32, int64, error) {
	var lastErr error
	for attempt := 1; attempt <= embedMaxAttempts; attempt++ {
		vecs, tokens, err := c.embedBatch(ctx, batch, model)
		if err == nil {
			return vecs, tokens, nil
		}
		lastErr = err
		if !isRetryableEmbedError(err) {
			return nil, 0, err
		}
		if attempt == embedMaxAttempts {
			break
		}
		delay := time.Duration(math.Min(
			float64(embedRetryBaseDelay)*math.Pow(2, float64(attempt-1)),
			float64(embedRetryMaxDelay),
		))
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, 0, lastErr
}

// embedBatch is a single OpenAI API call. Splits openai-go's typed
// response into ([][]float32, tokens) so the rest of the worker
// doesn't need to know about the SDK shape.
func (c *Client) embedBatch(ctx context.Context, batch []string, model EmbeddingModel) ([][]float32, int64, error) {
	resp, err := c.sdk.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: batch,
		},
	})
	if err != nil {
		return nil, 0, err
	}

	vecs := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		vec := make([]float32, len(d.Embedding))
		for j, f := range d.Embedding {
			vec[j] = float32(f)
		}
		vecs[i] = vec
	}
	return vecs, resp.Usage.TotalTokens, nil
}

// isRetryableEmbedError says whether an OpenAI error should be
// retried. 429 (rate limit) and 5xx are retryable; everything else
// (auth, validation, ctx cancel) is terminal.
func isRetryableEmbedError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests ||
			apiErr.StatusCode >= 500
	}
	// Network-level errors (connection reset, EOF) are also worth a
	// retry. errors.As matched no typed API error -- treat as transient.
	return true
}

// logEmbedCost emits the structured cost log for an Embed call.
// Mirrors logCost's shape but encodes input_tokens only (embeddings
// have no output tokens) and adds a per-model cost_usd lookup. An
// unmapped model logs cost_usd=0 + a warning attr so dashboards
// don't silently report a wrong number.
func (c *Client) logEmbedCost(userID uuid.UUID, feature Feature, model EmbeddingModel, requestID string, tokens int64, err error) {
	cost, knownPrice := usdForEmbedding(model, tokens)
	attrs := []slog.Attr{
		slog.String("feature", string(feature)),
		slog.String("user_id", userID.String()),
		slog.String("model", string(model)),
		slog.String("request_id", requestID),
		slog.Int64("input_tokens", tokens),
		slog.Float64("cost_usd", cost),
		slog.Bool("price_known", knownPrice),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		c.logger.LogAttrs(context.Background(), slog.LevelError, "ai_embed_failed", attrs...)
		return
	}
	c.logger.LogAttrs(context.Background(), slog.LevelInfo, "ai_embed_complete", attrs...)
}

// recordEmbedUsage persists an ai_usage row when a recorder is
// wired, so the daily quota service sees embedding spend alongside
// chat / edit. Skipped silently when c.recorder is nil (test path
// or partial wiring).
//
// Bounded with recordUsageTimeout so a slow / unavailable Postgres
// can't stall the worker indefinitely. The cost log line above
// remains the audit-only sibling of this row, so a dropped record
// here is recoverable from logs.
func (c *Client) recordEmbedUsage(userID uuid.UUID, feature Feature, model EmbeddingModel, requestID string, tokens int64) {
	if c.recorder == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), recordUsageTimeout)
	defer cancel()
	if err := c.recorder.RecordUsage(ctx, UsageRecord{
		UserID:    userID,
		Feature:   feature,
		Model:     Model(model),
		Usage:     Usage{InputTokens: tokens},
		RequestID: requestID,
	}); err != nil {
		c.logger.LogAttrs(context.Background(), slog.LevelError,
			"ai_embed_record_failed",
			slog.String("user_id", userID.String()),
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
		)
	}
}
