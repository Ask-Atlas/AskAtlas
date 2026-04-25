-- Queries for the ai_usage cost ledger (ASK-214).
--
-- Two consumers:
--
--   * The quota middleware counts today's spend per (user, feature)
--     before dispatching to OpenAI; rejecting with 429 on overflow.
--   * The cost-log hook in internal/ai/client.go writes one row per
--     completed (or cancelled) request so partial usage is still
--     attributed to the user who initiated it.
--
-- Index `idx_ai_usage_user_feature_created` is the supporting index
-- for the COUNT below.

-- name: InsertAIUsage :one
-- Records one billable AI request. Called from the cost-log hook
-- after the upstream stream terminates (success, error, OR ctx
-- cancellation). Returns the inserted row so callers can correlate
-- via id if needed.
INSERT INTO ai_usage (
  user_id,
  feature,
  model,
  input_tokens,
  output_tokens,
  cache_read_tokens,
  cache_write_tokens,
  request_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: CountAIUsageSince :one
-- Returns the number of ai_usage rows for the given user + feature
-- since `since`. The middleware passes UTC midnight as `since` to
-- enforce a daily quota; future per-hour quotas reuse the same
-- query with a different bound.
SELECT COUNT(*)::bigint AS count
FROM ai_usage
WHERE user_id = $1
  AND feature = $2
  AND created_at >= $3;
