-- Queries for the AI edit audit table (ASK-215).
--
-- Two consumers:
--
--   * The /api/study-guides/{id}/ai/edit handler INSERTs one row
--     after streaming the replacement to the client. Includes the
--     full input + output text so future eval / replay is possible.
--   * The /api/study-guides/{id}/ai/edits/{edit_id} PATCH handler
--     SETs accepted + accepted_at when the user resolves the diff.

-- name: InsertStudyGuideEdit :one
-- Persisted after a successful Stream completes. selection_start /
-- selection_end are character offsets in the rendered article body
-- (TipTap coordinate space, same one the selection toolbar uses).
-- instruction is the user's free-form rewrite directive.
INSERT INTO study_guide_edits (
  study_guide_id,
  user_id,
  instruction,
  selection_start,
  selection_end,
  original_span,
  replacement,
  model,
  input_tokens,
  output_tokens
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetStudyGuideEdit :one
-- Look up by (edit_id, study_guide_id) -- the path scoping prevents
-- a caller from updating an edit row attached to a different guide
-- they happen to know the id of. Caller still applies user-id
-- ownership in the service layer.
SELECT *
FROM study_guide_edits
WHERE id = $1
  AND study_guide_id = $2;

-- name: UpdateStudyGuideEditAcceptance :one
-- Records the user's accept/reject decision. Idempotent: re-PATCH
-- with the same value is a no-op. Once-only is NOT enforced -- the
-- frontend can correct a misclick within the editor session by
-- PATCH-ing again. Last write wins.
UPDATE study_guide_edits
SET
  accepted    = $3,
  accepted_at = $4
WHERE id = $1
  AND study_guide_id = $2
  AND user_id = $5
RETURNING *;
