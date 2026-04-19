-- Practice session queries (ASK-128 start/resume).
--
-- The StartPracticeSession service flow stitches these together:
--   1. CheckQuizLiveForSession  -- 404 dispatch (quiz live + parent live)
--   2. DeleteStaleIncompleteSessions -- hard-delete this user's
--      incomplete session for this quiz if started_at > 7 days ago
--   3. FindIncompleteSession -- if found, hydrate + return 200 (resume)
--   4. CountQuizQuestions (reused from ASK-115) -- snapshot count
--   5. Inside InTx:
--        a. InsertPracticeSessionIfAbsent -- ON CONFLICT DO NOTHING
--           backed by the partial unique index added in
--           20260419083647_add_practice_sessions_partial_unique_index.
--           Returns 0 rows on race -> service falls back to step 3.
--        b. SnapshotQuizQuestions -- bulk insert practice_session_questions
--           rows from quiz_questions
--   6. ListSessionAnswers -- for the resume response (empty array on
--      newly-created sessions because no answers exist yet)

-- name: CheckQuizLiveForSession :one
-- Returns TRUE when both the quiz row AND its parent study guide are
-- live (deleted_at IS NULL on both). The service uses this as the
-- 404 gate before any writes -- a soft-deleted quiz, a quiz under a
-- soft-deleted guide, and a missing quiz all return FALSE here, which
-- the service maps to a single 404 response so the caller cannot
-- distinguish them (info-leak prevention).
SELECT EXISTS (
  SELECT 1
  FROM quizzes q
  JOIN study_guides sg ON sg.id = q.study_guide_id
  WHERE q.id = sqlc.arg(quiz_id)::uuid
    AND q.deleted_at IS NULL
    AND sg.deleted_at IS NULL
) AS exists;

-- name: DeleteStaleIncompleteSessions :exec
-- Hard-deletes this user's incomplete session for this quiz when it
-- has been sitting around for more than 7 days. CASCADE deletes the
-- attached practice_session_questions and practice_answers rows.
--
-- Scoped per (user_id, quiz_id): we don't want a global cleanup job
-- here -- the spec wants stale-cleanup to run on the start-session
-- path so a user explicitly choosing to start fresh sees a clean
-- slate without waiting for a background job. Other users' stale
-- sessions on this same quiz are left alone (they'll be cleaned up
-- the next time THEY hit start-session).
--
-- The partial unique index makes this idempotent -- at most one row
-- can match the WHERE clause, so DELETE is a no-op when there's no
-- stale session, and a single-row DELETE otherwise.
DELETE FROM practice_sessions
WHERE user_id = sqlc.arg(user_id)::uuid
  AND quiz_id = sqlc.arg(quiz_id)::uuid
  AND completed_at IS NULL
  AND started_at < now() - interval '7 days';

-- name: FindIncompleteSession :one
-- Resume probe -- returns this user's current in-progress session for
-- the quiz, if any. The partial unique index on
-- (user_id, quiz_id) WHERE completed_at IS NULL guarantees AT MOST
-- one row matches, so LIMIT 1 is belt-and-suspenders.
--
-- Returns sql.ErrNoRows when no incomplete session exists; the
-- service treats that as the "create new" signal.
SELECT id, quiz_id, started_at, completed_at, total_questions, correct_answers
FROM practice_sessions
WHERE user_id = sqlc.arg(user_id)::uuid
  AND quiz_id = sqlc.arg(quiz_id)::uuid
  AND completed_at IS NULL
LIMIT 1;

-- name: InsertPracticeSessionIfAbsent :one
-- Race-safe insert backed by the partial unique index from migration
-- 20260419083647. ON CONFLICT DO NOTHING means a concurrent start by
-- the same user on the same quiz collapses to "no row inserted"
-- (sqlc returns sql.ErrNoRows for the :one annotation), and the
-- service catches that and falls back to FindIncompleteSession.
--
-- The losing request never touches the quiz_questions snapshot --
-- the winner already created it.
INSERT INTO practice_sessions (user_id, quiz_id, total_questions)
VALUES (
  sqlc.arg(user_id)::uuid,
  sqlc.arg(quiz_id)::uuid,
  sqlc.arg(total_questions)::integer
)
ON CONFLICT (user_id, quiz_id) WHERE completed_at IS NULL DO NOTHING
RETURNING id, quiz_id, started_at, completed_at, total_questions, correct_answers;

-- name: SnapshotQuizQuestions :exec
-- Freezes the quiz's CURRENT question set into the new session's
-- practice_session_questions rows. Runs inside the same tx as
-- InsertPracticeSessionIfAbsent so a partial snapshot can never be
-- observed by another reader -- either the entire session + all
-- questions exist, or neither does.
--
-- Subsequent edits to the quiz do not retroactively affect this
-- snapshot:
--   * New questions added later are not in this snapshot.
--   * Deleted questions trigger ON DELETE SET NULL on
--     practice_session_questions.question_id -- the snapshot row
--     persists with question_id = NULL so total_questions stays
--     stable.
INSERT INTO practice_session_questions (session_id, question_id, sort_order)
SELECT sqlc.arg(session_id)::uuid, id, sort_order
FROM quiz_questions
WHERE quiz_id = sqlc.arg(quiz_id)::uuid;

-- name: ListSessionAnswers :many
-- All answers submitted in a session, ordered by answered_at ASC so
-- the response renders them in the order the user produced them.
-- Used by the resume path to populate the `answers` array in
-- PracticeSessionResponse. Returns an empty slice (not nil) for
-- freshly-created sessions where no answers exist yet -- the
-- mapper renders that as `[]` rather than `null` per spec.
SELECT question_id, user_answer, is_correct, verified, answered_at
FROM practice_answers
WHERE session_id = sqlc.arg(session_id)::uuid
ORDER BY answered_at ASC;
