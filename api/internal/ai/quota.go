package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
)

// Quotas is the per-feature daily request cap. A zero value means
// "no cap" -- useful for tests and the safety-valve "other" bucket
// that catches Features we forgot to cap explicitly.
type Quotas map[Feature]int

// DefaultQuotas reflects the locked must-haves: editing is the
// flagship feature so it gets the largest envelope; quiz generation
// is expensive so capped tightly. Tweak per env via env vars (see
// QuotasFromEnv in this package).
var DefaultQuotas = Quotas{
	FeaturePing:         100,
	FeatureEdit:         50,
	FeatureGroundedEdit: 30,
	FeatureQA:           100,
	FeatureQuiz:         10,
	FeatureRefSuggest:   20,
}

// QuotaQuerier is the slice of db.Querier the quota service needs.
// Defined where it's used so tests can pass a generated fake without
// bringing in the full sqlc surface.
type QuotaQuerier interface {
	InsertAIUsage(ctx context.Context, arg db.InsertAIUsageParams) (db.AiUsage, error)
	CountAIUsageSince(ctx context.Context, arg db.CountAIUsageSinceParams) (int64, error)
}

// QuotaService gates AI requests against a per-user daily limit and
// records every request that proceeds. Safe for concurrent use.
//
// Implements the UsageRecorder contract so Client.Stream / Complete
// can write a row from the cost-log hook (including on cancellation
// so partial usage still bills).
type QuotaService struct {
	queries QuotaQuerier
	quotas  Quotas
	now     func() time.Time
}

// NewQuotaService wires the service over a sqlc Queries handle. Pass
// nil quotas to use DefaultQuotas. Tests inject WithClock to freeze
// the day boundary.
func NewQuotaService(queries QuotaQuerier, quotas Quotas, opts ...QuotaOption) *QuotaService {
	if quotas == nil {
		quotas = DefaultQuotas
	}
	s := &QuotaService{
		queries: queries,
		quotas:  quotas,
		now:     time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// QuotaOption tunes a QuotaService. Functional-options pattern keeps
// the constructor signature stable while letting tests override the
// clock without piling on parameters.
type QuotaOption func(*QuotaService)

// WithClock injects a deterministic time source. Tests use it to
// freeze "today" at a known instant.
func WithClock(now func() time.Time) QuotaOption {
	return func(s *QuotaService) { s.now = now }
}

// CheckAndReserve verifies the (user, feature) pair is under the
// daily cap. Returns *QuotaExceededError when over -- callers map
// that into a 429. Returns nil for features with no configured cap.
//
// "Reserve" is aspirational: we don't actually decrement an in-mem
// counter. The day-boundary COUNT against ai_usage is the source of
// truth, so concurrent requests can race past the cap by at most the
// concurrency-N. Acceptable for an MVP daily quota; revisit if we
// need strict per-second enforcement.
func (s *QuotaService) CheckAndReserve(ctx context.Context, userID uuid.UUID, feature Feature) error {
	limit, ok := s.quotas[feature]
	if !ok || limit <= 0 {
		return nil
	}
	midnight := utcMidnight(s.now())
	used, err := s.queries.CountAIUsageSince(ctx, db.CountAIUsageSinceParams{
		UserID:    utils.UUID(userID),
		Feature:   db.AiFeature(string(feature)),
		CreatedAt: utils.Timestamptz(&midnight),
	})
	if err != nil {
		return fmt.Errorf("ai: count usage: %w", err)
	}
	if used >= int64(limit) {
		return &QuotaExceededError{
			Feature: feature,
			Used:    used,
			Limit:   limit,
			ResetAt: midnight.Add(24 * time.Hour),
		}
	}
	return nil
}

// RecordUsage persists a single ai_usage row. Called from the AI
// client's cost-log hook on every stream termination -- success,
// error, or cancellation. Idempotency: not enforced; legitimate SDK
// retries can produce duplicate rows under the same request_id.
// Counting wins over deduplication for billing accuracy.
func (s *QuotaService) RecordUsage(ctx context.Context, rec UsageRecord) error {
	_, err := s.queries.InsertAIUsage(ctx, db.InsertAIUsageParams{
		UserID:           utils.UUID(rec.UserID),
		Feature:          db.AiFeature(string(rec.Feature)),
		Model:            string(rec.Model),
		InputTokens:      rec.Usage.InputTokens,
		OutputTokens:     rec.Usage.OutputTokens,
		CacheReadTokens:  rec.Usage.CacheReadTokens,
		CacheWriteTokens: rec.Usage.CacheWriteTokens,
		RequestID:        rec.RequestID,
	})
	if err != nil {
		return fmt.Errorf("ai: insert usage: %w", err)
	}
	return nil
}

// QuotaExceededError is the sentinel CheckAndReserve returns when a
// user has hit their daily cap. Carries the data the 429 envelope
// needs: feature, current usage, configured cap, and the absolute
// reset time (next UTC midnight) so middleware can compute
// Retry-After.
type QuotaExceededError struct {
	Feature Feature
	Used    int64
	Limit   int
	ResetAt time.Time
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf(
		"ai quota exceeded for feature %q: %d/%d (resets %s)",
		e.Feature, e.Used, e.Limit, e.ResetAt.UTC().Format(time.RFC3339),
	)
}

// IsQuotaExceeded narrows an arbitrary error to *QuotaExceededError.
// Middleware uses this instead of errors.As-with-pointer-pointer at
// the call site.
func IsQuotaExceeded(err error) (*QuotaExceededError, bool) {
	var qe *QuotaExceededError
	if errors.As(err, &qe) {
		return qe, true
	}
	return nil, false
}

// UsageRecord is what callers (AI client cost-log hook) emit to the
// QuotaService once a stream finishes. Mirrors the fields the cost
// log already populates.
type UsageRecord struct {
	UserID    uuid.UUID
	Feature   Feature
	Model     Model
	Usage     Usage
	RequestID string
}

// UsageRecorder is the narrow interface Client uses to persist
// usage. Defined here for documentation; QuotaService satisfies it.
type UsageRecorder interface {
	RecordUsage(ctx context.Context, rec UsageRecord) error
}

// QuotasFromEnv builds a Quotas map from AI_QUOTA_*_PER_DAY env
// vars, falling back to DefaultQuotas for any missing/unparseable
// entry. Lets ops dial individual features up/down per environment
// without redeploying code -- e.g. raise edit quota during a demo
// week, drop quiz quota when GPT-4 pricing spikes.
func QuotasFromEnv() Quotas {
	out := make(Quotas, len(DefaultQuotas))
	for feature, fallback := range DefaultQuotas {
		out[feature] = envQuota(feature, fallback)
	}
	return out
}

// envQuota reads AI_QUOTA_<UPPER_FEATURE>_PER_DAY with the given
// fallback. Negative values are clamped to zero (= uncapped) so a
// typo can only relax the quota, never silently invert the sign.
func envQuota(feature Feature, fallback int) int {
	key := "AI_QUOTA_" + featureEnvSegment(feature) + "_PER_DAY"
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

// featureEnvSegment uppercases a Feature for an env var key. Wraps
// strings.ToUpper so the conversion lives next to the only call site
// that depends on it.
func featureEnvSegment(feature Feature) string {
	out := make([]byte, 0, len(feature))
	for i := 0; i < len(feature); i++ {
		c := feature[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		out = append(out, c)
	}
	return string(out)
}

// utcMidnight returns the most recent UTC midnight on or before t.
// All quota windows are UTC-aligned so a user can't game the cap by
// hitting the API around their local midnight.
func utcMidnight(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
