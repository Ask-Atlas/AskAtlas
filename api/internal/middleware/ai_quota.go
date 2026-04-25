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

// FeatureForPath maps a (method, path) pair to a Feature when the
// route initiates an AI request that should be quota-gated and
// timeout-exempt for SSE. Returns (_, false) for non-AI routes -- the
// AIQuota middleware passes them through unchanged.
//
// Splitting AI detection out of the middleware lets a single mapper
// power both the quota gate AND the timeout-skipper in main.go, so
// the two can never disagree about whether a route is "AI".
//
// Only POST routes that hit the model count as AI. PATCH/GET on the
// audit/usage tables don't burn tokens; they pass through.
//
// Update this mapper every time a new AI endpoint ships. The list is
// intentionally explicit (not regex/glob) so the bill column on
// every endpoint is visible in one place at code-review time.
type FeatureForPath func(method, path string) (ai.Feature, bool)

// DefaultFeatureForPath is the production mapping. Add new endpoints
// here as they ship.
func DefaultFeatureForPath(method, path string) (ai.Feature, bool) {
	if method != http.MethodPost {
		return "", false
	}
	switch {
	case path == "/api/ai/ping":
		return ai.FeaturePing, true
	case isStudyGuideAIPath(path, "edit"):
		return ai.FeatureEdit, true
	// Future endpoints register here:
	//   case isStudyGuideAIPath(path, "grounded-edit"): return ai.FeatureGroundedEdit, true
	//   case isStudyGuideAIPath(path, "qa"):            return ai.FeatureQA, true
	//   case isStudyGuideAIPath(path, "quiz"):          return ai.FeatureQuiz, true
	//   case isStudyGuideAIPath(path, "refs/suggest"):  return ai.FeatureRefSuggest, true
	default:
		return "", false
	}
}

// isStudyGuideAIPath matches /api/study-guides/<id>/ai/<suffix> with
// exactly one path segment for <id>. We don't fully validate the
// UUID -- the OpenAPI validator already rejects malformed IDs before
// middleware runs. Just need enough specificity to avoid hits on a
// hypothetical /api/study-guides/<id>/ai-history kind of route.
func isStudyGuideAIPath(path, suffix string) bool {
	const prefix = "/api/study-guides/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	want := "/ai/" + suffix
	if !strings.HasSuffix(rest, want) {
		return false
	}
	id := rest[:len(rest)-len(want)]
	return id != "" && !strings.ContainsRune(id, '/')
}

// IsAIRoute is the public form of the mapper's bool. The timeout-
// skipper in main.go calls this to align bypass decisions with the
// quota gate -- a route is AI for both or for neither.
func IsAIRoute(mapper FeatureForPath, method, path string) bool {
	if mapper == nil {
		mapper = DefaultFeatureForPath
	}
	_, ok := mapper(method, path)
	return ok
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
// Non-AI routes (per the mapper) pass through untouched. Mount once
// at the /api group; the mapper decides which subset to gate.
func AIQuota(gate QuotaGate, mapper FeatureForPath) func(http.Handler) http.Handler {
	if mapper == nil {
		mapper = DefaultFeatureForPath
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			feature, isAI := mapper(r.Method, r.URL.Path)
			if !isAI {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := authctx.UserIDFromContext(r.Context())
			if !ok {
				apperrors.RespondWithError(w, apperrors.NewUnauthorized())
				return
			}

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
