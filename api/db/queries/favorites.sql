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
