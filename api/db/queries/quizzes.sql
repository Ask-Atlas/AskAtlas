-- Quizzes write + read queries (ASK-150 / ASK-136).
--
-- The create flow is wrapped in a single InTx in the service layer:
-- InsertQuiz -> N x InsertQuizQuestion -> M x InsertQuizAnswerOption.
-- A failure at any step rolls everything back, so a partial quiz can
-- never be observed by another reader.
--
-- The post-insert hydration runs OUTSIDE the transaction (commit
-- happens first) using GetQuizDetail + ListQuizQuestionsByQuiz +
-- ListQuizAnswerOptionsByQuiz. The two-list fan-out matches the
-- studyguides detail pattern -- mapping options back onto questions
-- happens in Go because pgx returns flat rowsets and the question
-- count is small (<=100 per quiz).
--
-- Privacy floor on the creator payload mirrors studyguides: id +
-- first_name + last_name only. No email, no clerk_id.

-- name: GuideExistsAndLiveForQuizzes :one
-- Live-presence probe for the create-quiz endpoint. Returns TRUE
-- only when the guide row exists AND is not soft-deleted. The
-- studyguides package has an identical query (GuideExistsAndLive);
-- duplicated here so the quizzes service can stay decoupled from
-- the studyguides Repository interface (sqlc generates queriers
-- per package -- both call the same row but live in different
-- generated method tables).
SELECT EXISTS (
  SELECT 1
  FROM study_guides
  WHERE id = sqlc.arg(id)::uuid
    AND deleted_at IS NULL
) AS exists;

-- name: InsertQuiz :one
-- Insert a new quiz row. Returns the columns the service needs to
-- build the QuizDetailResponse without an extra round trip on the
-- write side -- the read-side hydration still happens via
-- GetQuizDetail because the creator's first_name + last_name come
-- from a join to users (and would inflate this RETURNING clause).
INSERT INTO quizzes (study_guide_id, creator_id, title, description)
VALUES (
  sqlc.arg(study_guide_id)::uuid,
  sqlc.arg(creator_id)::uuid,
  sqlc.arg(title)::text,
  sqlc.narg(description)::text
)
RETURNING id, created_at, updated_at;

-- name: InsertQuizQuestion :one
-- Insert a single question row. reference_answer is only meaningful
-- for `freeform` questions; the service passes NULL for the other
-- two types. sort_order is required (the service sets a stable
-- value -- either the user-supplied integer or the array index).
INSERT INTO quiz_questions (
  quiz_id, type, question_text, hint,
  feedback_correct, feedback_incorrect, reference_answer, sort_order
)
VALUES (
  sqlc.arg(quiz_id)::uuid,
  sqlc.arg(type)::question_type,
  sqlc.arg(question_text)::text,
  sqlc.narg(hint)::text,
  sqlc.narg(feedback_correct)::text,
  sqlc.narg(feedback_incorrect)::text,
  sqlc.narg(reference_answer)::text,
  sqlc.arg(sort_order)::integer
)
RETURNING id;

-- name: InsertQuizAnswerOption :exec
-- Insert one option row. The service has already validated that
-- exactly one option per MCQ has is_correct=true; for true-false
-- questions the service synthesises two options (`True` + `False`)
-- with the matching is_correct flag.
INSERT INTO quiz_answer_options (question_id, text, is_correct, sort_order)
VALUES (
  sqlc.arg(question_id)::uuid,
  sqlc.arg(text)::text,
  sqlc.arg(is_correct)::boolean,
  sqlc.arg(sort_order)::integer
);

-- name: GetQuizDetail :one
-- Load the quiz row + privacy-floor creator info for the detail
-- payload. The study guide is NOT joined back -- the caller already
-- knows the study_guide_id (it's in the URL on POST and on the
-- quiz row itself). Excludes soft-deleted quizzes (deleted_at IS
-- NULL), soft-deleted creators (u.deleted_at IS NULL), and
-- soft-deleted parent guides (sg.deleted_at IS NULL) so a hydration
-- that races with a parent-cascade soft-delete reports 'not found'
-- rather than rendering an orphaned quiz.
SELECT
  q.id, q.study_guide_id, q.title, q.description,
  q.created_at, q.updated_at,
  u.id          AS creator_id,
  u.first_name  AS creator_first_name,
  u.last_name   AS creator_last_name
FROM quizzes q
JOIN users u ON u.id = q.creator_id
JOIN study_guides sg ON sg.id = q.study_guide_id
WHERE q.id = sqlc.arg(id)::uuid
  AND q.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND sg.deleted_at IS NULL;

-- name: ListQuizQuestionsByQuiz :many
-- All questions for a quiz, ordered by sort_order then id (the id
-- tiebreaker keeps the response deterministic when two questions
-- happen to share a sort_order -- the spec doesn't enforce
-- uniqueness on sort_order). Returns reference_answer so the
-- mapper can emit it as `correct_answer` on freeform questions.
SELECT
  id, type, question_text, hint,
  feedback_correct, feedback_incorrect, reference_answer, sort_order
FROM quiz_questions
WHERE quiz_id = sqlc.arg(quiz_id)::uuid
ORDER BY sort_order ASC, id ASC;

-- name: GetQuizByIDForUpdate :one
-- Locked SELECT used at the start of DeleteQuiz (ASK-102) and
-- UpdateQuiz (ASK-153). SELECT FOR UPDATE prevents two concurrent
-- mutators from racing on the same row -- one wins with 204/200,
-- the other sees the post-mutation state in its tx snapshot and
-- returns 404. Filters NOTHING -- the service inspects deleted_at
-- + creator_id to choose 404 vs 403 vs proceed (mirrors
-- studyguides.GetStudyGuideByIDForUpdate).
SELECT id, creator_id, deleted_at
FROM quizzes
WHERE id = sqlc.arg(id)::uuid
FOR UPDATE;

-- name: GetQuizForUpdateWithParentStatus :one
-- Locked SELECT used by UpdateQuiz (ASK-153). Like
-- GetQuizByIDForUpdate but also joins to the parent study guide
-- so the service can enforce the "parent guide must be live"
-- gate (AC6) inside the same transaction without a second round
-- trip. Locks ONLY the quiz row (FOR UPDATE OF q) -- the parent
-- guide is read-only at this site and we don't want to hold a
-- write lock on it for the duration of the quiz update.
--
-- Filters NOTHING -- the service inspects the four fields to
-- choose 404 (missing quiz / deleted quiz / deleted guide) vs
-- 403 (not creator) vs proceed.
SELECT q.id, q.creator_id, q.deleted_at,
       sg.deleted_at AS guide_deleted_at
FROM quizzes q
JOIN study_guides sg ON sg.id = q.study_guide_id
WHERE q.id = sqlc.arg(id)::uuid
FOR UPDATE OF q;

-- name: UpdateQuiz :exec
-- Partial update for ASK-153. The title column uses the standard
-- COALESCE(narg, current) pattern -- nil means "leave alone".
--
-- Description uses a CASE expression because COALESCE can't
-- distinguish "field absent in the request" from "field
-- explicitly null in the request" (both encode as SQL NULL once
-- they reach the query layer). The handler resolves the
-- tri-state, sets `clear_description` true on an explicit
-- `description: null` body, and the service drives both args
-- accordingly:
--   * clear_description=false                 -> column unchanged
--   * clear_description=true, description=NULL -> column cleared
--   * clear_description=true, description="x"  -> column set to "x"
--
-- The service is responsible for 404 / 403 gating before this
-- query runs (via GetQuizForUpdateWithParentStatus). updated_at
-- is bumped to now() on every successful call -- the at-least-
-- one-field rule is enforced in Go so this query is never
-- reached for an empty PATCH.
UPDATE quizzes
SET
  title       = COALESCE(sqlc.narg(title)::text, title),
  description = CASE WHEN sqlc.arg(clear_description)::bool
                     THEN sqlc.narg(description)::text
                     ELSE description END,
  updated_at  = now()
WHERE id = sqlc.arg(id)::uuid;

-- name: SoftDeleteQuiz :exec
-- Set deleted_at = now() on the quiz. The service has already
-- verified the row exists, isn't already deleted, and the viewer
-- is the creator -- so this is a blind UPDATE. No cascade: practice
-- sessions, questions, and answer options are preserved per the
-- ASK-102 spec ("preserve historical practice data; the quiz
-- simply becomes invisible to list/detail endpoints").
UPDATE quizzes
SET deleted_at = now()
WHERE id = sqlc.arg(id)::uuid;

-- name: ListQuizzesByStudyGuide :many
-- Lists every non-soft-deleted quiz attached to a study guide
-- (ASK-136). Each row carries the privacy-floor creator payload
-- (id + first_name + last_name only -- mirrors the studyguides
-- surface) plus a server-computed question_count via LEFT JOIN +
-- COUNT, so a quiz with zero questions still surfaces with a 0
-- count rather than being silently dropped by an INNER JOIN.
--
-- Soft-delete invariants:
--   * q.deleted_at IS NULL  -- excludes soft-deleted quizzes
--                              (AC3: studyguide with mixed live +
--                              deleted quizzes only returns the live
--                              ones)
--   * u.deleted_at IS NULL  -- excludes quizzes whose creator's user
--                              record was soft-deleted (matches the
--                              ASK-143 convention used by
--                              GetStudyGuideDetail)
--
-- The parent study guide's deleted_at is checked separately by the
-- service via GuideExistsAndLiveForQuizzes BEFORE this query runs --
-- a missing or soft-deleted guide returns 404 even when this query
-- would have returned an empty array. Keeps the 404 vs 200-empty
-- distinction crisp.
--
-- Order: created_at DESC with id DESC as the deterministic
-- tiebreaker (the spec calls for "newest first"). The id tiebreaker
-- prevents a flaky test on a Postgres that happens to insert two
-- rows in the same microsecond.
SELECT
  q.id, q.title, q.description, q.created_at, q.updated_at,
  u.id          AS creator_id,
  u.first_name  AS creator_first_name,
  u.last_name   AS creator_last_name,
  COUNT(qq.id)::bigint AS question_count
FROM quizzes q
JOIN users u ON u.id = q.creator_id
LEFT JOIN quiz_questions qq ON qq.quiz_id = q.id
WHERE q.study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND q.deleted_at IS NULL
  AND u.deleted_at IS NULL
GROUP BY q.id, u.id
ORDER BY q.created_at DESC, q.id DESC;

-- name: ListQuizAnswerOptionsByQuiz :many
-- All answer options for every question in a quiz, ordered by
-- question_id then sort_order then id. The mapper groups by
-- question_id in Go to attach options to their parent question.
-- The triple-key ordering keeps the option list deterministic.
SELECT
  o.id, o.question_id, o.text, o.is_correct, o.sort_order
FROM quiz_answer_options o
JOIN quiz_questions qq ON qq.id = o.question_id
WHERE qq.quiz_id = sqlc.arg(quiz_id)::uuid
ORDER BY o.question_id ASC, o.sort_order ASC, o.id ASC;
