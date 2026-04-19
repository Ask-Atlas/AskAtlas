-- Study guide list queries (ASK-104).
--
-- Every ListStudyGuides* variant uses the same CTE structure so the
-- per-row aggregates (vote_score, is_recommended, quiz_count) are
-- computed once and can be referenced from the outer WHERE clause
-- (e.g. for the score-sorted cursor predicate). The CTE pattern also
-- keeps the 8 named variants near-identical apart from ORDER BY + the
-- cursor predicate, which makes future maintenance (e.g. adding a new
-- sort field) a mechanical edit.
--
-- Soft-delete invariants enforced everywhere:
--   * sg.deleted_at IS NULL       — excludes guides marked for deletion
--   * u.deleted_at IS NULL        — excludes guides authored by a
--                                   soft-deleted user; matches the
--                                   convention established by ASK-143's
--                                   section roster (a soft-deleted user
--                                   disappears from public surfaces)
--   * quizzes.deleted_at IS NULL  — quiz_count excludes deleted quizzes
--
-- Privacy floor on the creator payload: only id + first_name + last_name
-- are selected. No email, no clerk_id -- same rule as
-- SectionMemberResponse in ASK-143.

-- name: InsertStudyGuide :one
-- Insert a new guide and return all the columns the service needs to
-- construct the StudyGuideDetail response without an extra round trip.
-- The course preflight (in service.go) gates on AssertCourseExists so
-- the FK violation is unreachable in normal flow; the FK still acts
-- as a backstop if a course is hard-deleted between preflight + insert.
INSERT INTO study_guides (course_id, creator_id, title, description, content, tags)
VALUES (
  sqlc.arg(course_id)::uuid,
  sqlc.arg(creator_id)::uuid,
  sqlc.arg(title)::text,
  sqlc.narg(description)::text,
  sqlc.narg(content)::text,
  sqlc.arg(tags)::text[]
)
RETURNING id, view_count, created_at, updated_at;

-- name: GetStudyGuideByIDForUpdate :one
-- Locked SELECT used at the start of DeleteStudyGuide. SELECT FOR
-- UPDATE prevents concurrent deletes from racing on the same guide
-- (one wins with 204, the other sees the row already-deleted in its
-- transaction's snapshot and returns 404). Filters NOTHING -- the
-- service inspects deleted_at + creator_id to choose 404 vs 403 vs
-- proceed.
SELECT id, creator_id, deleted_at
FROM study_guides
WHERE id = sqlc.arg(id)::uuid
FOR UPDATE;

-- name: SoftDeleteStudyGuide :exec
-- Set deleted_at = now() on the guide. The service has already
-- verified the row exists, isn't already deleted, and the viewer is
-- the creator -- so this is a blind UPDATE. The DeleteStudyGuide
-- transaction wraps this + SoftDeleteQuizzesForGuide.
UPDATE study_guides
SET deleted_at = now()
WHERE id = sqlc.arg(id)::uuid;

-- name: SoftDeleteQuizzesForGuide :exec
-- Application-level cascade: soft-delete every non-deleted quiz on
-- the guide. WHERE deleted_at IS NULL preserves the deleted_at
-- timestamp on quizzes that were already soft-deleted before the
-- guide was -- the spec explicitly requires that an already-deleted
-- quiz's deleted_at is NOT updated by this cascade.
UPDATE quizzes
SET deleted_at = now()
WHERE study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND deleted_at IS NULL;

-- name: GetStudyGuideDetail :one
-- The detail endpoint's main query (ASK-114). Returns the guide's own
-- columns + a compact course payload + a compact creator payload
-- + two inline aggregates as subqueries:
--   * vote_score    -- SUM(up/down votes)
--   * is_recommended -- EXISTS in study_guide_recommendations
--
-- The viewer's own vote (user_vote) ships in a separate query
-- (GetUserVoteForGuide) because sqlc does not infer nullable output
-- columns from LEFT JOIN / subquery expressions on enum-typed columns
-- -- it reads the schema's NOT NULL constraint and types the output
-- non-nullable. An extra round trip is cheaper than fighting sqlc's
-- type inference; the PRD's "batching as separate queries" guidance
-- explicitly allows it.
--
-- Soft-delete invariants:
--   * sg.deleted_at IS NULL  -- excludes deleted guides (→ 404)
--   * u.deleted_at IS NULL   -- creator must be live (ASK-143 convention)
--
-- Privacy floor: no email, no clerk_id. Creator exposes only
-- id/first_name/last_name.
SELECT
  sg.id, sg.title, sg.description, sg.content, sg.tags,
  sg.view_count, sg.created_at, sg.updated_at,
  c.id           AS course_id,
  c.department   AS course_department,
  c.number       AS course_number,
  c.title        AS course_title,
  u.id           AS creator_id,
  u.first_name   AS creator_first_name,
  u.last_name    AS creator_last_name,
  COALESCE((
    SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
    FROM study_guide_votes
    WHERE study_guide_id = sg.id
  ), 0)::bigint AS vote_score,
  EXISTS (
    SELECT 1 FROM study_guide_recommendations
    WHERE study_guide_id = sg.id
  ) AS is_recommended
FROM study_guides sg
JOIN courses c ON c.id = sg.course_id
JOIN users   u ON u.id = sg.creator_id
WHERE sg.id = sqlc.arg(id)::uuid
  AND sg.deleted_at IS NULL
  AND u.deleted_at IS NULL;

-- name: GetUserVoteForGuide :one
-- Returns the viewer's own vote on the guide, or sql.ErrNoRows when
-- the viewer has not voted. The service maps ErrNoRows to a nil
-- user_vote in the response (JSON null, not omitted).
SELECT vote
FROM study_guide_votes
WHERE study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND user_id = sqlc.arg(viewer_id)::uuid;

-- name: ListGuideRecommenders :many
-- Recommenders list for the guide detail payload. Same privacy floor
-- as CreatorSummary -- id + first_name + last_name only. Excludes
-- recommenders whose user record is soft-deleted.
SELECT u.id, u.first_name, u.last_name
FROM study_guide_recommendations sgr
JOIN users u ON u.id = sgr.recommended_by
WHERE sgr.study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND u.deleted_at IS NULL
ORDER BY sgr.created_at ASC, u.id ASC;

-- name: ListGuideQuizzesWithQuestionCount :many
-- Non-deleted quizzes for the guide + question_count per quiz. The
-- LEFT JOIN ensures quizzes with zero questions still appear with
-- question_count = 0.
SELECT
  q.id, q.title,
  COUNT(qq.id)::bigint AS question_count
FROM quizzes q
LEFT JOIN quiz_questions qq ON qq.quiz_id = q.id
WHERE q.study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND q.deleted_at IS NULL
GROUP BY q.id
ORDER BY q.created_at ASC, q.id ASC;

-- name: ListGuideResources :many
-- Attached resources for the guide detail payload. No creator info
-- in the SELECT list -- the caller doesn't need to know who attached
-- the resource.
SELECT r.id, r.title, r.url, r.type, r.description, r.created_at
FROM study_guide_resources sgr
JOIN resources r ON r.id = sgr.resource_id
WHERE sgr.study_guide_id = sqlc.arg(study_guide_id)::uuid
ORDER BY sgr.created_at ASC, r.id ASC;

-- name: ListGuideFiles :many
-- Attached files for the guide detail payload. Privacy floor: no
-- user_id, no s3_key, no checksum. The file list shows only what a
-- viewer needs to see: what's attached, what type, and how big.
--
-- Filters f.status = 'complete' so files mid-upload (pending) or
-- failed don't surface in the guide detail -- a frontend that tried
-- to download such a file would get a broken link. Only successfully
-- uploaded files are visible to non-owners; the upload author's own
-- file list (via the files endpoints) shows all statuses so they can
-- retry or remove.
SELECT f.id, f.name, f.mime_type, f.size
FROM study_guide_files sgf
JOIN files f ON f.id = sgf.file_id
WHERE sgf.study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND f.status = 'complete'
ORDER BY sgf.created_at ASC, f.id ASC;

-- name: CourseExistsForGuides :one
-- Single-row probe used by the list handler to disambiguate "course
-- missing" (404) from "course exists but has no guides" (200 empty
-- array). Separate from courses.CourseExists only because sqlc-generated
-- queriers are per-file; the predicate is identical.
SELECT EXISTS (
  SELECT 1 FROM courses WHERE id = sqlc.arg(id)::uuid
) AS exists;

-- name: ListStudyGuidesScoreDesc :many
-- Default sort. Multi-column keyset on (vote_score, view_count,
-- updated_at, id) -- each column after vote_score is a tiebreaker for
-- the previous one; id is the final strict-total-order tiebreaker.
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*)
      FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_vote_score)::bigint IS NULL
  OR (vote_score, view_count, updated_at, id) < (
    sqlc.narg(cursor_vote_score)::bigint,
    sqlc.narg(cursor_view_count)::bigint,
    sqlc.narg(cursor_updated_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY vote_score DESC, view_count DESC, updated_at DESC, id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesScoreAsc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_vote_score)::bigint IS NULL
  OR (vote_score, view_count, updated_at, id) > (
    sqlc.narg(cursor_vote_score)::bigint,
    sqlc.narg(cursor_view_count)::bigint,
    sqlc.narg(cursor_updated_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY vote_score ASC, view_count ASC, updated_at ASC, id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesViewsDesc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_view_count)::bigint IS NULL
  OR (view_count, updated_at, id) < (
    sqlc.narg(cursor_view_count)::bigint,
    sqlc.narg(cursor_updated_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY view_count DESC, updated_at DESC, id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesViewsAsc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_view_count)::bigint IS NULL
  OR (view_count, updated_at, id) > (
    sqlc.narg(cursor_view_count)::bigint,
    sqlc.narg(cursor_updated_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY view_count ASC, updated_at ASC, id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesNewestDesc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_created_at)::timestamptz IS NULL
  OR (created_at, id) < (
    sqlc.narg(cursor_created_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesNewestAsc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_created_at)::timestamptz IS NULL
  OR (created_at, id) > (
    sqlc.narg(cursor_created_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY created_at ASC, id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesUpdatedDesc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_updated_at)::timestamptz IS NULL
  OR (updated_at, id) < (
    sqlc.narg(cursor_updated_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY updated_at DESC, id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListStudyGuidesUpdatedAsc :many
WITH scored AS (
  SELECT
    sg.id, sg.title, sg.description, sg.tags, sg.course_id,
    sg.view_count, sg.created_at, sg.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name,
    COALESCE((
      SELECT SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END)::bigint
      FROM study_guide_votes
      WHERE study_guide_id = sg.id
    ), 0)::bigint AS vote_score,
    EXISTS (
      SELECT 1 FROM study_guide_recommendations
      WHERE study_guide_id = sg.id
    ) AS is_recommended,
    (
      SELECT COUNT(*) FROM quizzes
      WHERE study_guide_id = sg.id AND deleted_at IS NULL
    )::bigint AS quiz_count
  FROM study_guides sg
  JOIN users u ON u.id = sg.creator_id
  WHERE sg.course_id = sqlc.arg(course_id)::uuid
    AND sg.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND (
      sqlc.narg(q)::text IS NULL
      OR sg.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      OR EXISTS (
        SELECT 1 FROM unnest(sg.tags) AS t
        WHERE t ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
      )
    )
    AND (sqlc.narg(tags)::text[] IS NULL OR sg.tags @> sqlc.narg(tags)::text[])
)
SELECT id, title, description, tags, course_id, view_count,
       created_at, updated_at,
       creator_id, creator_first_name, creator_last_name,
       vote_score, is_recommended, quiz_count
FROM scored
WHERE (
  sqlc.narg(cursor_updated_at)::timestamptz IS NULL
  OR (updated_at, id) > (
    sqlc.narg(cursor_updated_at)::timestamptz,
    sqlc.narg(cursor_id)::uuid
  )
)
ORDER BY updated_at ASC, id ASC
LIMIT sqlc.arg(page_limit);

-- name: GuideExistsAndLive :one
-- Live-presence probe used by both vote endpoints. Returns TRUE only
-- when the guide row exists AND is not soft-deleted. The vote service
-- gates on this before the upsert/delete so a missing-or-deleted
-- guide returns 404 with a clear message rather than e.g. trampling
-- through to the SQL layer and surfacing a generic FK error.
SELECT EXISTS (
  SELECT 1
  FROM study_guides
  WHERE id = sqlc.arg(id)::uuid
    AND deleted_at IS NULL
) AS exists;

-- name: UpsertStudyGuideVote :exec
-- Cast or change a vote (ASK-139). Inserts a new (user_id,
-- study_guide_id, vote) row when the viewer has not voted, or
-- updates the existing row's vote when the direction changes. Same-
-- direction re-submits hit the WHERE clause on the DO UPDATE branch
-- and become a true no-op (no row touched, no trigger fired,
-- updated_at preserved). The (user_id, study_guide_id) PK from the
-- schema is what makes ON CONFLICT resolve correctly.
INSERT INTO study_guide_votes (user_id, study_guide_id, vote)
VALUES (
  sqlc.arg(user_id)::uuid,
  sqlc.arg(study_guide_id)::uuid,
  sqlc.arg(vote)::vote_direction
)
ON CONFLICT (user_id, study_guide_id) DO UPDATE
  SET vote = EXCLUDED.vote,
      updated_at = now()
  WHERE study_guide_votes.vote != EXCLUDED.vote;

-- name: ComputeGuideVoteScore :one
-- Recomputes the guide's vote_score from study_guide_votes. Returned
-- as int64 to match the wire shape on CastVoteResponse. Run after
-- the upsert in the same logical request so the response reflects the
-- post-mutation state.
SELECT COALESCE(
  SUM(CASE WHEN vote = 'up' THEN 1 ELSE -1 END),
  0
)::bigint AS vote_score
FROM study_guide_votes
WHERE study_guide_id = sqlc.arg(study_guide_id)::uuid;

-- name: DeleteStudyGuideVote :execrows
-- Hard-delete the (viewer, guide) vote row (ASK-141). Returns the
-- rows-affected count so the service can distinguish "no existing
-- vote" (0 rows -> 404 'Vote not found') from a successful delete
-- (1 row -> 204). The guide-existence check happens BEFORE this
-- runs in the service so a missing guide doesn't leak through as
-- "vote not found".
DELETE FROM study_guide_votes
WHERE user_id = sqlc.arg(user_id)::uuid
  AND study_guide_id = sqlc.arg(study_guide_id)::uuid;

-- name: ViewerCanRecommendForGuide :one
-- Combined live-presence + role-gate probe for the recommend
-- endpoints (ASK-147 + ASK-101). Returns one row when the viewer
-- holds instructor or ta role in AT LEAST ONE section of the guide's
-- course AND the guide is live (not soft-deleted).
--
-- Returns three booleans so the service can distinguish 404 from
-- 403 with a single round trip:
--   * guide_exists  -- guide row present AND deleted_at IS NULL
--   * has_role      -- viewer is instructor/ta in some section
--                      of the guide's course (ignored if guide
--                      doesn't exist)
--
-- Combining the two checks into a single query (rather than two
-- sequential calls) keeps the recommend hot path at one DB round
-- trip for the gate; the actual insert/delete is the second.
WITH guide AS (
  SELECT id, course_id
  FROM study_guides
  WHERE id = sqlc.arg(study_guide_id)::uuid
    AND deleted_at IS NULL
)
SELECT
  EXISTS (SELECT 1 FROM guide) AS guide_exists,
  EXISTS (
    SELECT 1
    FROM course_members cm
    JOIN course_sections cs ON cs.id = cm.section_id
    WHERE cs.course_id = (SELECT course_id FROM guide)
      AND cm.user_id = sqlc.arg(viewer_id)::uuid
      AND cm.role IN ('instructor', 'ta')
  ) AS has_role;

-- name: InsertStudyGuideRecommendation :one
-- Inserts the (study_guide_id, recommended_by) row and returns the
-- server-generated created_at so the response can ship the timestamp
-- without a follow-up SELECT. The (study_guide_id, recommended_by)
-- PK from the schema makes a duplicate insert raise unique_violation
-- (Postgres SQLSTATE 23505), which the service catches and maps to
-- apperrors.ErrConflict (409).
INSERT INTO study_guide_recommendations (study_guide_id, recommended_by)
VALUES (
  sqlc.arg(study_guide_id)::uuid,
  sqlc.arg(recommended_by)::uuid
)
RETURNING created_at;

-- name: DeleteStudyGuideRecommendation :execrows
-- Hard-delete the (viewer, guide) recommendation row. Returns the
-- rows-affected count so the service can distinguish "viewer never
-- recommended this guide" (0 rows -> 404 'Recommendation not found')
-- from a successful delete (1 row -> 204). The guide-existence +
-- role gate runs FIRST in the service so 'Study guide not found' /
-- 403 win over 'Recommendation not found' when applicable.
DELETE FROM study_guide_recommendations
WHERE study_guide_id = sqlc.arg(study_guide_id)::uuid
  AND recommended_by = sqlc.arg(recommended_by)::uuid;
