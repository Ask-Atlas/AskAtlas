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
-- has been sitting around longer than the caller-supplied stale
-- threshold. CASCADE deletes the attached practice_session_questions
-- and practice_answers rows.
--
-- The threshold is a Go-side constant (sessions.StaleSessionAge,
-- currently 7 days) passed in as seconds so the policy lives in
-- exactly one place. Multiplying by `interval '1 second'` lets us
-- pass a plain integer instead of a Postgres interval value, which
-- keeps the sqlc-generated Go signature simple.
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
  AND started_at < now() - (sqlc.arg(stale_threshold_seconds)::bigint * interval '1 second');

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
-- total_questions is intentionally NOT supplied here -- the column
-- defaults to 0 and is set to the authoritative snapshot row count
-- by SnapshotQuizQuestionsAndUpdateCount in the same tx. This
-- avoids the race window that existed when total_questions was
-- pre-computed via CountQuizQuestions and could disagree with the
-- actual snapshot row count under concurrent quiz edits at
-- READ COMMITTED isolation (gemini + copilot PR #153 feedback).
--
-- The losing request never touches the quiz_questions snapshot --
-- the winner already created it.
INSERT INTO practice_sessions (user_id, quiz_id)
VALUES (
  sqlc.arg(user_id)::uuid,
  sqlc.arg(quiz_id)::uuid
)
ON CONFLICT (user_id, quiz_id) WHERE completed_at IS NULL DO NOTHING
RETURNING id, quiz_id, started_at, completed_at, total_questions, correct_answers;

-- name: SnapshotQuizQuestionsAndUpdateCount :one
-- Atomic snapshot + total_questions fix-up. Inserts one
-- practice_session_questions row per current quiz_questions row in
-- a single CTE, then updates the session's total_questions to the
-- count of rows actually inserted. Returns the new total_questions
-- so the caller can sync its in-memory session state.
--
-- The single-statement CTE eliminates the race that existed in the
-- prior two-step approach (CountQuizQuestions sets total_questions
-- on insert, then a separate SnapshotQuizQuestions writes the
-- snapshot). Under Postgres' default READ COMMITTED isolation each
-- statement gets its own snapshot, so a concurrent quiz edit
-- between count and snapshot could leave the two out of sync.
-- Within a single CTE statement, both reads share the same
-- statement-level snapshot, so the count and the snapshot are
-- guaranteed identical (gemini + copilot PR #153 feedback).
--
-- Subsequent edits to the quiz do not retroactively affect this
-- snapshot:
--   * New questions added AFTER this statement are not in the snapshot.
--   * Deleted questions trigger ON DELETE SET NULL on
--     practice_session_questions.question_id -- the snapshot row
--     persists with question_id = NULL so total_questions stays
--     stable.
WITH inserted AS (
  INSERT INTO practice_session_questions (session_id, question_id, sort_order)
  SELECT sqlc.arg(session_id)::uuid, id, sort_order
  FROM quiz_questions
  WHERE quiz_id = sqlc.arg(quiz_id)::uuid
  RETURNING 1
)
UPDATE practice_sessions
SET total_questions = (SELECT count(*)::integer FROM inserted)
WHERE id = sqlc.arg(session_id)::uuid
RETURNING total_questions;

-- name: GetSessionForAnswerSubmission :one
-- Locks the session row and returns ownership + completion state
-- for the answer-submit endpoint (ASK-137). FOR UPDATE serializes
-- against a concurrent SessionComplete (ASK-140 future) so the
-- answer either commits before the completion (recorded) or
-- after (rejected with 409). Filters NOTHING -- the service
-- inspects user_id + completed_at to choose 404 / 403 / 409 /
-- proceed.
SELECT id, user_id, completed_at
FROM practice_sessions
WHERE id = sqlc.arg(id)::uuid
FOR UPDATE;

-- name: CheckQuestionInSessionSnapshot :one
-- Returns TRUE when the question_id is part of this session's
-- frozen practice_session_questions snapshot. Used by ASK-137 to
-- enforce the spec's "question must be in this session's
-- snapshot" rule (400 otherwise -- the user can only answer
-- questions that were in the quiz at session-start time).
SELECT EXISTS (
  SELECT 1 FROM practice_session_questions
  WHERE session_id = sqlc.arg(session_id)::uuid
    AND question_id = sqlc.arg(question_id)::uuid
) AS exists;

-- name: GetCorrectOptionText :one
-- Returns the text of the option marked is_correct for an MCQ or
-- TF question. Used by ASK-137 to compare against user_answer:
--   * multiple-choice -- exact string equality with this text
--   * true-false -- this text is "True" or "False" (the canonical
--     labels written by the create-quiz path); the service maps
--     it to a boolean and compares against the user's parsed
--     "true"/"false" input.
-- Returns sql.ErrNoRows if no option is marked correct, which
-- the service treats as a data-integrity 500 (write-side
-- validation should have prevented it).
SELECT text
FROM quiz_answer_options
WHERE question_id = sqlc.arg(question_id)::uuid
  AND is_correct = true
LIMIT 1;

-- name: InsertPracticeAnswer :one
-- Records the user's answer with the backend-determined
-- is_correct + verified flags (ASK-137). The unique constraint
-- uq_practice_answers_session_question on (session_id,
-- question_id) catches duplicate submissions; the service
-- detects the unique-violation pgconn error and surfaces a
-- typed 400 with details {"question_id": "already answered"}.
--
-- Returns the persisted columns so the handler can render the
-- PracticeAnswerResponse without a re-fetch.
INSERT INTO practice_answers (session_id, question_id, user_answer, is_correct, verified)
VALUES (
  sqlc.arg(session_id)::uuid,
  sqlc.arg(question_id)::uuid,
  sqlc.arg(user_answer)::text,
  sqlc.arg(is_correct)::boolean,
  sqlc.arg(verified)::boolean
)
RETURNING question_id, user_answer, is_correct, verified, answered_at;

-- name: IncrementSessionCorrectAnswers :exec
-- Bumps practice_sessions.correct_answers by 1. Called only when
-- the inserted answer was correct (ASK-137 AC8). Wrapped in the
-- same tx as InsertPracticeAnswer so a failure rolls back the
-- answer row too -- the counter and the underlying answer can
-- never disagree.
UPDATE practice_sessions
SET correct_answers = correct_answers + 1
WHERE id = sqlc.arg(id)::uuid;

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
