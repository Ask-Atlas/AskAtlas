-- name: ListFileFavorites :many
-- Per-viewer file favorites (ASK-151). Joins file_favorites onto
-- files so the row carries name + mime_type without a follow-up
-- GET. Filters f.deletion_status IS NULL to exclude files in any
-- deletion lifecycle (mirrors every other files.sql query).
--
-- Ordered by ff.created_at DESC, file_id DESC for a strict total
-- order over the favorites table (created_at alone isn't unique
-- under load, file_id is the tiebreaker).
--
-- Index used: PK (user_id, file_id). The PK isn't a covering index
-- for ORDER BY created_at -- file_favorites in the existing schema
-- has no idx_file_favorites_user_created index. At MVP scale (a
-- single user's favorites) the planner does an in-memory sort over
-- the user's rows after a PK lookup, which is acceptable. If
-- favorite counts grow into thousands per user, add the index.
SELECT
  ff.file_id    AS file_id,
  ff.created_at AS favorited_at,
  f.name        AS file_name,
  f.mime_type   AS file_mime_type
FROM file_favorites ff
JOIN files f ON f.id = ff.file_id
WHERE ff.user_id = sqlc.arg(viewer_id)
  AND f.deletion_status IS NULL
ORDER BY ff.created_at DESC, ff.file_id DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: ListStudyGuideFavorites :many
-- Per-viewer study-guide favorites (ASK-151). Joins study_guides
-- for title and the parent course's department + number so the
-- sidebar can render a "CPTS 322 -- <title>" label without a
-- follow-up request. Filters sg.deleted_at IS NULL because the
-- favorites widget should never surface the user's own deleted
-- guides.
--
-- Index used: idx_study_guide_favorites_user_created
-- (user_id, created_at, study_guide_id).
SELECT
  sgf.study_guide_id AS study_guide_id,
  sgf.created_at     AS favorited_at,
  sg.title           AS study_guide_title,
  c.department       AS course_department,
  c.number           AS course_number
FROM study_guide_favorites sgf
JOIN study_guides sg ON sg.id = sgf.study_guide_id
JOIN courses c       ON c.id = sg.course_id
WHERE sgf.user_id = sqlc.arg(viewer_id)
  AND sg.deleted_at IS NULL
ORDER BY sgf.created_at DESC, sgf.study_guide_id DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CheckFileExists :one
-- Existence probe used by POST /api/files/{file_id}/favorite (ASK-130).
-- Returns sql.ErrNoRows when the file is missing or in any deletion
-- lifecycle state -- both map to 404 per the spec ("file not found
-- or soft-deleted"). The favorite toggle only requires the file to
-- exist; favoriting is intentionally permission-less per the spec
-- ("any authenticated user can favorite any non-deleted file").
SELECT 1
FROM files
WHERE id = sqlc.arg(file_id)::uuid
  AND deletion_status IS NULL;

-- name: CheckStudyGuideExists :one
-- Existence probe used by POST /api/me/study-guides/{study_guide_id}/favorite
-- (ASK-156). Returns sql.ErrNoRows when missing or soft-deleted.
-- Same rationale as CheckFileExists -- favoriting is permission-less.
SELECT 1
FROM study_guides
WHERE id = sqlc.arg(study_guide_id)::uuid
  AND deleted_at IS NULL;

-- name: CheckCourseExists :one
-- Existence probe used by POST /api/me/courses/{course_id}/favorite
-- (ASK-157). Courses do not support soft-delete, so existence is the
-- only condition. Returns sql.ErrNoRows when missing.
SELECT 1
FROM courses
WHERE id = sqlc.arg(course_id)::uuid;

-- name: ToggleFileFavorite :one
-- Toggles file_favorites for (user, file) (ASK-130). The CTE deletes
-- any existing row first, then inserts only when the delete found
-- nothing -- one round trip, no read-modify-write race. The spec
-- treats concurrent toggles as deterministic per request: PostgreSQL
-- serializes them via row locks on file_favorites' PK.
--
-- Returns:
--   * favorited     = true  + favorited_at = NOW()  when the row was
--                     newly inserted (state flipped to favorited).
--   * favorited     = false + favorited_at = NULL   when the row was
--                     deleted (state flipped to unfavorited).
--
-- Existence of the parent file is gated by CheckFileExists upstream
-- so this query trusts its inputs; it never touches the files table.
WITH deleted AS (
  DELETE FROM file_favorites
  WHERE user_id = sqlc.arg(user_id)::uuid
    AND file_id = sqlc.arg(file_id)::uuid
  RETURNING 1
), inserted AS (
  INSERT INTO file_favorites (user_id, file_id)
  SELECT sqlc.arg(user_id)::uuid, sqlc.arg(file_id)::uuid
  WHERE NOT EXISTS (SELECT 1 FROM deleted)
  RETURNING created_at
)
SELECT
  EXISTS (SELECT 1 FROM inserted)::boolean AS favorited,
  (SELECT created_at FROM inserted)        AS favorited_at;

-- name: ToggleStudyGuideFavorite :one
-- Same shape as ToggleFileFavorite, against study_guide_favorites
-- (ASK-156).
WITH deleted AS (
  DELETE FROM study_guide_favorites
  WHERE user_id = sqlc.arg(user_id)::uuid
    AND study_guide_id = sqlc.arg(study_guide_id)::uuid
  RETURNING 1
), inserted AS (
  INSERT INTO study_guide_favorites (user_id, study_guide_id)
  SELECT sqlc.arg(user_id)::uuid, sqlc.arg(study_guide_id)::uuid
  WHERE NOT EXISTS (SELECT 1 FROM deleted)
  RETURNING created_at
)
SELECT
  EXISTS (SELECT 1 FROM inserted)::boolean AS favorited,
  (SELECT created_at FROM inserted)        AS favorited_at;

-- name: ToggleCourseFavorite :one
-- Same shape as ToggleFileFavorite, against course_favorites
-- (ASK-157).
WITH deleted AS (
  DELETE FROM course_favorites
  WHERE user_id = sqlc.arg(user_id)::uuid
    AND course_id = sqlc.arg(course_id)::uuid
  RETURNING 1
), inserted AS (
  INSERT INTO course_favorites (user_id, course_id)
  SELECT sqlc.arg(user_id)::uuid, sqlc.arg(course_id)::uuid
  WHERE NOT EXISTS (SELECT 1 FROM deleted)
  RETURNING created_at
)
SELECT
  EXISTS (SELECT 1 FROM inserted)::boolean AS favorited,
  (SELECT created_at FROM inserted)        AS favorited_at;

-- name: ListCourseFavorites :many
-- Per-viewer course favorites (ASK-151). Joins courses for the
-- (department, number, title) triple. No soft-delete filter --
-- the courses table does not support soft-delete.
--
-- Index used: idx_course_favorites_user_created
-- (user_id, created_at, course_id).
SELECT
  cf.course_id  AS course_id,
  cf.created_at AS favorited_at,
  c.department  AS course_department,
  c.number      AS course_number,
  c.title       AS course_title
FROM course_favorites cf
JOIN courses c ON c.id = cf.course_id
WHERE cf.user_id = sqlc.arg(viewer_id)
ORDER BY cf.created_at DESC, cf.course_id DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);
