-- name: ListRecentFiles :many
-- Per-viewer recent files (ASK-145). Joins file_last_viewed onto files
-- so the row carries the display payload (name + mime_type) the
-- sidebar needs without a follow-up GET. Filters
-- f.deletion_status IS NULL to exclude files in any deletion lifecycle
-- (pending_deletion or deleted) -- mirrors the same predicate every
-- other files.sql query uses.
--
-- Index used: idx_file_last_viewed_user_viewed_file
-- (user_id, viewed_at, file_id).
SELECT
  lv.file_id   AS file_id,
  lv.viewed_at AS viewed_at,
  f.name       AS file_name,
  f.mime_type  AS file_mime_type
FROM file_last_viewed lv
JOIN files f ON f.id = lv.file_id
WHERE lv.user_id = sqlc.arg(viewer_id)
  AND f.deletion_status IS NULL
ORDER BY lv.viewed_at DESC
LIMIT sqlc.arg(page_limit);

-- name: ListRecentStudyGuides :many
-- Per-viewer recent study guides (ASK-145). Joins study_guides for
-- title and the parent course's department + number so the sidebar
-- can render "CPTS 322 -- Binary Trees Cheat Sheet" without a
-- follow-up request. Filters sg.deleted_at IS NULL because the
-- recents widget should never surface the user's own deleted guides
-- (the my-guides endpoint is the place to view/restore those).
--
-- Index used: idx_study_guide_last_viewed_user_viewed
-- (user_id, viewed_at, study_guide_id).
SELECT
  lv.study_guide_id AS study_guide_id,
  lv.viewed_at      AS viewed_at,
  sg.title          AS study_guide_title,
  c.department      AS course_department,
  c.number          AS course_number
FROM study_guide_last_viewed lv
JOIN study_guides sg ON sg.id = lv.study_guide_id
JOIN courses c       ON c.id = sg.course_id
WHERE lv.user_id = sqlc.arg(viewer_id)
  AND sg.deleted_at IS NULL
ORDER BY lv.viewed_at DESC
LIMIT sqlc.arg(page_limit);

-- name: ListRecentCourses :many
-- Per-viewer recent courses (ASK-145). Joins courses for the
-- (department, number, title) triple. No soft-delete filter --
-- the courses table does not support soft-delete.
--
-- Index used: idx_course_last_viewed_user_viewed
-- (user_id, viewed_at, course_id).
SELECT
  lv.course_id AS course_id,
  lv.viewed_at AS viewed_at,
  c.department AS course_department,
  c.number     AS course_number,
  c.title      AS course_title
FROM course_last_viewed lv
JOIN courses c ON c.id = lv.course_id
WHERE lv.user_id = sqlc.arg(viewer_id)
ORDER BY lv.viewed_at DESC
LIMIT sqlc.arg(page_limit);
