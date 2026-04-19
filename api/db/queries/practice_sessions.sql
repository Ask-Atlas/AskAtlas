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

-- name: GetSessionByID :one
-- Reads a session row by id. Used by GetPracticeSession (ASK-152)
-- to render the session detail. Unlike LockSessionForCompletion
-- this does NOT FOR UPDATE -- it's a pure read so concurrent
-- writers (SubmitAnswer, CompleteSession) shouldn't block.
--
-- Returns ALL session fields so the service can build the wire
-- response (including the nullable completed_at the score
-- calculator gates on).
--
-- No parent quiz / study_guide deletion check: sessions are
-- historical data and remain accessible even after the parent
-- quiz or guide is soft-deleted (ASK-152 spec AC6 + technical
-- note: "sessions must remain accessible even after the quiz is
-- removed").
SELECT id, user_id, quiz_id, started_at, completed_at, total_questions, correct_answers
FROM practice_sessions
WHERE id = sqlc.arg(id)::uuid;

-- name: LockSessionForCompletion :one
-- Locks the session row and returns ALL fields the
-- CompleteSession endpoint (ASK-140) needs to assemble its
-- response. FOR UPDATE serializes against a concurrent
-- SubmitAnswer (ASK-137 also FOR UPDATEs the session row), so
-- the spec's "answer-vs-complete race -> first commit wins"
-- semantics fall out naturally:
--   * answer wins -> complete sees correct_answers updated
--   * complete wins -> answer's locked SELECT sees completed_at
--     set and returns 409
--
-- Returns sql.ErrNoRows when the session doesn't exist; the
-- service maps that to 404. The presence of completed_at on the
-- returned row drives the 409-vs-proceed decision.
SELECT id, user_id, quiz_id, started_at, completed_at, total_questions, correct_answers
FROM practice_sessions
WHERE id = sqlc.arg(id)::uuid
FOR UPDATE;

-- name: MarkSessionCompleted :one
-- Sets completed_at = now() and returns the timestamp the row
-- now carries. The service uses the returned timestamp to
-- assemble the response without a re-fetch (the rest of the
-- session fields were captured by LockSessionForCompletion in
-- the same tx, so they don't need to round-trip again).
--
-- This is a blind UPDATE: the service has already verified
-- ownership + completed_at IS NULL inside the same tx via
-- LockSessionForCompletion + the FOR UPDATE row lock. By the
-- time this runs, the only legitimate outcome is "row updated".
UPDATE practice_sessions
SET completed_at = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING completed_at;

-- name: DeleteSessionByID :execrows
-- Hard-deletes a practice session by id (ASK-144). The CASCADE
-- foreign keys on practice_session_questions and practice_answers
-- ensure the snapshot rows and answer rows are removed in the
-- same statement.
--
-- Blind by id only -- the service has already verified ownership
-- + completed_at IS NULL inside the same tx via
-- LockSessionForCompletion + the FOR UPDATE row lock. By the
-- time this runs, the only legitimate outcome is "row deleted".
-- :execrows lets the service double-check the rows-affected
-- count (defense-in-depth) and surface a 500 on the
-- vanishingly-rare 0-rows path (would mean another tx slipped
-- in and deleted between our lock and this DELETE -- which is
-- ruled out by the FOR UPDATE row lock under READ COMMITTED,
-- but the check is cheap and self-documenting).
DELETE FROM practice_sessions
WHERE id = sqlc.arg(id)::uuid;

-- name: ListUserSessionsForQuiz :many
-- Cursor-paginated keyset list of the authenticated user's
-- practice sessions for one quiz (ASK-149). Sorted by
-- (started_at DESC, id DESC) so newest attempts appear first;
-- id is the deterministic tie-breaker on the (vanishingly rare)
-- case of two sessions sharing started_at to the microsecond.
--
-- Filters:
--   * Scoped to (user_id, quiz_id). The handler/service
--     anchors user_id on the JWT so users cannot list each
--     other's sessions even if they spoof the path param.
--   * status_filter optional:
--       NULL       -- both active + completed (interleaved by started_at)
--       'active'   -- completed_at IS NULL  (in-progress only)
--       'completed'-- completed_at IS NOT NULL (finalised only)
--   * Keyset cursor (cursor_started_at + cursor_id) is the
--     started_at + id of the LAST row from the previous page.
--     Both nullable args MUST be set together; the service
--     decodes them as a pair from the opaque base64 cursor and
--     never sends one without the other. The query enforces
--     the pair invariant defensively in SQL (see the WHERE
--     clause below) so a half-set cursor surfaces as a clear
--     error rather than a mysteriously empty page (Postgres
--     tuple comparison against NULL evaluates to NULL, which
--     filters every row out). copilot PR #158 feedback.
--
-- Pagination: the service passes page_limit = caller_limit + 1
-- so it can detect has_more without an extra COUNT query --
-- if more than caller_limit rows come back, the extra row is
-- trimmed and has_more=true.
--
-- No parent quiz / study_guide deletion check inline here: the
-- service gates the call with CheckQuizLiveForSession before
-- invoking this query. A deleted parent surfaces as 404 before
-- this list query runs, so a stale "list" call against a
-- soft-deleted quiz cannot leak rows. (CheckQuizLiveForSession
-- is itself a DB query -- the "no DB hit" path is from the
-- application logic perspective, not the literal HTTP layer.)
SELECT id, started_at, completed_at, total_questions, correct_answers
FROM practice_sessions
WHERE user_id = sqlc.arg(user_id)::uuid
  AND quiz_id = sqlc.arg(quiz_id)::uuid
  AND (
    sqlc.narg(status_filter)::text IS NULL
    OR (sqlc.narg(status_filter)::text = 'active' AND completed_at IS NULL)
    OR (sqlc.narg(status_filter)::text = 'completed' AND completed_at IS NOT NULL)
  )
  AND (
    -- Half-set cursor guard: both args NULL (first page) or
    -- both args set (subsequent page). A half-set cursor is a
    -- caller bug -- collapse the row set to zero rows by
    -- comparing 1 = 0, then layer the real keyset condition
    -- on top of the both-set branch only. This makes a
    -- half-set cursor return zero rows deterministically
    -- instead of silently filtering everything out via
    -- NULL-tuple comparison.
    (sqlc.narg(cursor_started_at)::timestamptz IS NULL
     AND sqlc.narg(cursor_id)::uuid IS NULL)
    OR (sqlc.narg(cursor_started_at)::timestamptz IS NOT NULL
        AND sqlc.narg(cursor_id)::uuid IS NOT NULL
        AND (started_at, id) < (
          sqlc.narg(cursor_started_at)::timestamptz,
          sqlc.narg(cursor_id)::uuid
        ))
  )
ORDER BY started_at DESC, id DESC
LIMIT sqlc.arg(page_limit);

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
