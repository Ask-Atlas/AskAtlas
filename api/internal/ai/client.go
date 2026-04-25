package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// Client is the AskAtlas-facing wrapper around openai-go. It owns the
// API key, applies our model defaults, normalizes Chat-Completions
// streaming chunks into the package-level Event shape, writes the
// cost log, and (when configured) persists a row in ai_usage so the
// quota service can enforce daily caps.
//
// Construct via NewClient. The zero value is unusable.
type Client struct {
	sdk      openai.Client
	logger   *slog.Logger
	recorder UsageRecorder // may be nil -- see WithRecorder
}

// ClientOption tunes a Client at construction. Functional-options
// pattern keeps NewClient back-compat as we add optional dependencies
// (recorder, custom HTTP client, retries) over the lifetime of the
// AI epic.
type ClientOption func(*Client)

// WithRecorder wires a persistent usage recorder (typically the
// QuotaService) so every completed request writes one row to
// ai_usage. Without a recorder, Client still emits the structured
// cost log -- the row write is the addition.
func WithRecorder(r UsageRecorder) ClientOption {
	return func(c *Client) { c.recorder = r }
}

// NewClient builds a Client backed by the real OpenAI SDK. The
// caller is responsible for ensuring apiKey is non-empty -- config
// validation guards this at startup so we don't fail a request with
// "missing key" later.
func NewClient(apiKey string, logger *slog.Logger, opts ...ClientOption) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	c := &Client{
		sdk:    openai.NewClient(option.WithAPIKey(apiKey)),
		logger: logger,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Stream sends req to Claude and returns a channel of Events. The
// channel always closes -- on success after EventUsage + EventDone,
// on failure after EventError, on ctx cancellation without emitting
// further events.
//
// A request_id is allocated up-front so all log lines (and any future
// trace spans) share it. The cost log is written exactly once, when
// the upstream stream terminates, so partial usage from cancelled
// streams is still attributed.
func (c *Client) Stream(ctx context.Context, req StreamRequest) (<-chan Event, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	params := buildParams(req)

	requestID := uuid.NewString()

	out := make(chan Event, 16)
	go func() {
		defer close(out)
		usage := Usage{}

		stream := c.sdk.Chat.Completions.NewStreaming(ctx, params)

		for stream.Next() {
			chunk := stream.Current()

			// Text deltas appear in Choices[0].Delta.Content. Most
			// chunks emit one block; tool calls and finish-reason
			// chunks have empty Content and are skipped.
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta != "" {
					if !send(ctx, out, Event{Kind: EventDelta, Delta: delta}) {
						return
					}
				}
			}

			// The final chunk carries usage when stream_options
			// include_usage=true was requested. PromptTokens already
			// includes cache reads -- subtract to derive the un-
			// cached portion that bills at full rate.
			if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
				cached := chunk.Usage.PromptTokensDetails.CachedTokens
				usage.InputTokens = chunk.Usage.PromptTokens - cached
				usage.OutputTokens = chunk.Usage.CompletionTokens
				usage.CacheReadTokens = cached
				// OpenAI doesn't separately bill cache writes; field
				// kept for ASK-214 ledger schema parity.
				usage.CacheWriteTokens = 0
			}
		}

		if err := stream.Err(); err != nil {
			// Distinguish ctx cancellation (client disconnect) from a
			// real upstream failure -- only the latter should inflate
			// the error-rate dashboard. Cancellation still logs counts
			// so partial usage is attributed.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.logCancelled(req, requestID, usage)
				return
			}
			c.logCost(req, requestID, usage, err)
			_ = send(ctx, out, Event{Kind: EventError, Err: errors.New("upstream model error")})
			return
		}

		usageCopy := usage
		if !send(ctx, out, Event{Kind: EventUsage, Usage: &usageCopy}) {
			return
		}
		_ = send(ctx, out, Event{Kind: EventDone})

		c.logCost(req, requestID, usage, nil)
	}()

	return out, nil
}

// Complete sends req to Claude and waits for the full response, no
// streaming. Use this for short non-UI-bound tasks (classification,
// ranking, extraction) where token-by-token feedback isn't useful.
// Cost-log line is emitted on return, identical to Stream's.
func (c *Client) Complete(ctx context.Context, req StreamRequest) (CompleteResponse, error) {
	if err := req.validate(); err != nil {
		return CompleteResponse{}, err
	}
	requestID := uuid.NewString()
	params := buildParams(req)

	completion, err := c.sdk.Chat.Completions.New(ctx, params)
	if err != nil {
		c.logCost(req, requestID, Usage{}, err)
		return CompleteResponse{}, fmt.Errorf("ai: chat.completions.new: %w", err)
	}

	cached := completion.Usage.PromptTokensDetails.CachedTokens
	usage := Usage{
		InputTokens:      completion.Usage.PromptTokens - cached,
		OutputTokens:     completion.Usage.CompletionTokens,
		CacheReadTokens:  cached,
		CacheWriteTokens: 0,
	}

	var sb strings.Builder
	if len(completion.Choices) > 0 {
		sb.WriteString(completion.Choices[0].Message.Content)
	}
	c.logCost(req, requestID, usage, nil)
	return CompleteResponse{Text: sb.String(), Usage: usage}, nil
}

// logCost writes a single structured line per request; ASK-214's
// rate-limit middleware reads ai_usage to enforce daily caps, and
// the structured log is the audit-only sibling of that row.
// No secrets logged -- only counts + IDs.
func (c *Client) logCost(req StreamRequest, requestID string, usage Usage, err error) {
	model := req.Model
	if model == "" {
		model = ModelDefault
	}
	attrs := []slog.Attr{
		slog.String("feature", string(req.Feature)),
		slog.String("user_id", req.UserID.String()),
		slog.String("model", string(model)),
		slog.String("request_id", requestID),
		slog.Int64("input_tokens", usage.InputTokens),
		slog.Int64("output_tokens", usage.OutputTokens),
		slog.Int64("cache_read_tokens", usage.CacheReadTokens),
		slog.Int64("cache_write_tokens", usage.CacheWriteTokens),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		c.logger.LogAttrs(context.Background(), slog.LevelError, "ai_request_failed", attrs...)
		c.recordUsage(req, requestID, usage)
		return
	}
	c.logger.LogAttrs(context.Background(), slog.LevelInfo, "ai_request", attrs...)
	c.recordUsage(req, requestID, usage)
}

// recordUsage persists a row in ai_usage when a recorder is wired.
// Uses a detached context so a cancelled request still attributes
// its (partial) cost to the user -- the alternative would let
// abusive clients spam-cancel to dodge the daily quota.
//
// 5-second deadline is generous: Postgres INSERTs against the
// indexed (user_id, feature, created_at) shape complete in single-
// digit ms. Anything slower than that is a real DB problem and
// surfacing it as a log line is enough.
func (c *Client) recordUsage(req StreamRequest, requestID string, usage Usage) {
	if c.recorder == nil {
		return
	}
	model := req.Model
	if model == "" {
		model = ModelDefault
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rec := UsageRecord{
		UserID:    req.UserID,
		Feature:   req.Feature,
		Model:     model,
		Usage:     usage,
		RequestID: requestID,
	}
	if err := c.recorder.RecordUsage(ctx, rec); err != nil {
		c.logger.LogAttrs(context.Background(), slog.LevelError, "ai_usage_persist_failed",
			slog.String("request_id", requestID),
			slog.String("user_id", req.UserID.String()),
			slog.String("feature", string(req.Feature)),
			slog.String("error", err.Error()),
		)
	}
}

// logCancelled records partial usage for streams that ended because
// the caller went away. Same fields as logCost but at LevelInfo with
// `cancelled=true` so dashboards can filter.
func (c *Client) logCancelled(req StreamRequest, requestID string, usage Usage) {
	model := req.Model
	if model == "" {
		model = ModelDefault
	}
	c.logger.LogAttrs(context.Background(), slog.LevelInfo, "ai_request_cancelled",
		slog.String("feature", string(req.Feature)),
		slog.String("user_id", req.UserID.String()),
		slog.String("model", string(model)),
		slog.String("request_id", requestID),
		slog.Int64("input_tokens", usage.InputTokens),
		slog.Int64("output_tokens", usage.OutputTokens),
		slog.Int64("cache_read_tokens", usage.CacheReadTokens),
		slog.Int64("cache_write_tokens", usage.CacheWriteTokens),
		slog.Bool("cancelled", true),
	)
	// Cancelled requests still bill -- OpenAI charges input tokens
	// regardless of whether output completed. Skipping the row would
	// let abusive clients spam-cancel to dodge the daily quota.
	c.recordUsage(req, requestID, usage)
}

// send tries to deliver ev on out, but bails if the consumer or the
// context goes away. Returns true iff the event was delivered.
func send(ctx context.Context, out chan<- Event, ev Event) bool {
	select {
	case out <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}

func (req StreamRequest) validate() error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("ai: StreamRequest.UserID is required")
	}
	if req.Feature == "" {
		return fmt.Errorf("ai: StreamRequest.Feature is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("ai: StreamRequest.Messages must contain at least one turn")
	}
	for i, m := range req.Messages {
		if m.Role != RoleUser && m.Role != RoleAssistant {
			return fmt.Errorf("ai: Messages[%d].Role %q invalid", i, m.Role)
		}
		if len(m.Blocks) == 0 {
			return fmt.Errorf("ai: Messages[%d].Blocks is empty", i)
		}
	}
	return nil
}

// buildParams maps our StreamRequest onto the SDK's
// ChatCompletionNewParams. CacheControl markers are dropped --
// OpenAI auto-caches >= 1024-token prompts; the field on Block stays
// as a structural hint (callers should still place long stable
// content first).
func buildParams(req StreamRequest) openai.ChatCompletionNewParams {
	model := req.Model
	if model == "" {
		model = ModelDefault
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = MaxTokensDefault
	}
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	if sys := joinBlocks(req.System); sys != "" {
		messages = append(messages, openai.SystemMessage(sys))
	}
	for _, m := range req.Messages {
		text := joinBlocks(m.Blocks)
		switch m.Role {
		case RoleAssistant:
			messages = append(messages, openai.AssistantMessage(text))
		default:
			messages = append(messages, openai.UserMessage(text))
		}
	}
	return openai.ChatCompletionNewParams{
		Model:               model,
		MaxCompletionTokens: param.NewOpt(maxTokens),
		Messages:            messages,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: param.NewOpt(true),
		},
	}
}

// joinBlocks concatenates a Block slice into a single string. OpenAI
// Chat-Completions takes one content string per role; we glue blocks
// with double-newlines so retrieved chunks stay readable to the model.
func joinBlocks(blocks []Block) string {
	if len(blocks) == 0 {
		return ""
	}
	if len(blocks) == 1 {
		return blocks[0].Text
	}
	var sb strings.Builder
	for i, b := range blocks {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(b.Text)
	}
	return sb.String()
}
