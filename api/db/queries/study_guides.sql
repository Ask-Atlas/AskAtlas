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
