package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
)

// QuotaGate is the slice of ai.QuotaService the middleware needs.
// Defined here (where it's used) so tests can substitute a fake
// without depending on the full QuotaService surface.
type QuotaGate interface {
	CheckAndReserve(ctx context.Context, userID uuid.UUID, feature ai.Feature) error
}

// FeatureForPath maps a URL path to a Feature for quota attribution.
// Returns ai.FeaturePing as the safety fallback so we always charge
// _something_ if the path mapping drifts behind a new endpoint.
//
// Update this map every time a new /api/ai/* route ships. The list
// is intentionally explicit (not regex/glob) so the bill column on
// every endpoint is visible in one place at code-review time.
type FeatureForPath func(path string) ai.Feature

// DefaultFeatureForPath is the production mapping. Falls back to
// FeaturePing for any unrecognized /api/ai/* path -- the wrong
// bucket, but never zero, which keeps the ledger honest while we
// notice the missing entry.
func DefaultFeatureForPath(path string) ai.Feature {
	switch {
	case strings.HasPrefix(path, "/api/ai/ping"):
		return ai.FeaturePing
	// Future endpoints register here:
	//   case strings.HasPrefix(path, "/api/ai/edit"):           return ai.FeatureEdit
	//   case strings.HasPrefix(path, "/api/ai/grounded-edit"):  return ai.FeatureGroundedEdit
	//   case strings.HasPrefix(path, "/api/ai/qa"):             return ai.FeatureQA
	//   case strings.HasPrefix(path, "/api/ai/quiz"):           return ai.FeatureQuiz
	//   case strings.HasPrefix(path, "/api/ai/refs/suggest"):   return ai.FeatureRefSuggest
	default:
		return ai.FeaturePing
	}
}

// AIQuota returns a middleware that gates AI requests against a
// per-user daily quota. Pre-handler it calls
// QuotaGate.CheckAndReserve; on overrun it writes a 429 envelope
// with Retry-After and short-circuits without invoking next.
//
// The middleware does NOT write the ai_usage row -- that's the
// responsibility of the AI client's cost-log hook (see
// ai.Client.recordUsage), so partial usage on cancellation is still
// attributed.
//
// scope is the URL prefix (e.g. "/api/ai/") that this middleware
// gates. Requests outside the prefix pass through untouched, which
// lets the same handler chain serve both AI and non-AI routes
// without per-route wiring.
func AIQuota(gate QuotaGate, scope string, mapper FeatureForPath) func(http.Handler) http.Handler {
	if mapper == nil {
		mapper = DefaultFeatureForPath
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, scope) {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := authctx.UserIDFromContext(r.Context())
			if !ok {
				apperrors.RespondWithError(w, apperrors.NewUnauthorized())
				return
			}

			feature := mapper(r.URL.Path)
			err := gate.CheckAndReserve(r.Context(), userID, feature)
			if err == nil {
				next.ServeHTTP(w, r)
				return
			}

			if qe, isQuota := ai.IsQuotaExceeded(err); isQuota {
				writeQuotaExceeded(w, qe)
				return
			}

			slog.Error("AIQuota: CheckAndReserve failed", "error", err, "user_id", userID, "feature", feature)
			apperrors.RespondWithError(w, apperrors.NewInternalError())
		})
	}
}

// quotaExceededBody is the wire shape of the 429 response. Mirrors
// apperrors.AppError's code/status/message so frontend error
// handling can fall through to the same toast component, with the
// quota-specific fields tacked on for surfacing usage in UI.
type quotaExceededBody struct {
	Code              int        `json:"code"`
	Status            string     `json:"status"`
	Message           string     `json:"message"`
	Feature           ai.Feature `json:"feature"`
	Used              int64      `json:"used"`
	Limit             int        `json:"limit"`
	ResetAt           string     `json:"reset_at"`
	RetryAfterSeconds int        `json:"retry_after_seconds"`
}

func writeQuotaExceeded(w http.ResponseWriter, qe *ai.QuotaExceededError) {
	retryAfter := int(time.Until(qe.ResetAt).Seconds())
	if retryAfter < 1 {
		retryAfter = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	body := quotaExceededBody{
		Code:              http.StatusTooManyRequests,
		Status:            "QUOTA_EXCEEDED",
		Message:           "AI daily quota exceeded for this feature",
		Feature:           qe.Feature,
		Used:              qe.Used,
		Limit:             qe.Limit,
		ResetAt:           qe.ResetAt.UTC().Format(time.RFC3339),
		RetryAfterSeconds: retryAfter,
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("AIQuota: failed to encode 429 body", "error", err)
	}
}
