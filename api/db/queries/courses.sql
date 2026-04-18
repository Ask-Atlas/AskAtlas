-- All ListCourses* variants share the same WHERE template
-- (school_id, department, q filters + sort-specific cursor) and only
-- differ in ORDER BY direction and the cursor field shape. This mirrors
-- the per-sort-variant pattern in files.sql -- sqlc cannot parameterize
-- ORDER BY, so each direction gets its own named query.
--
-- The default sort (by department) uses a composite (department, number, id)
-- cursor because (department) alone is not unique; the other sort fields
-- use a simpler (field, id) cursor since the field is paired with the
-- primary key as a tiebreaker.

-- name: ListCoursesDepartmentAsc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_department)::text IS NULL
    OR (c.department, c.number, c.id) > (
      sqlc.narg(cursor_department)::text,
      sqlc.narg(cursor_number)::text,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY c.department ASC, c.number ASC, c.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesDepartmentDesc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_department)::text IS NULL
    OR (c.department, c.number, c.id) < (
      sqlc.narg(cursor_department)::text,
      sqlc.narg(cursor_number)::text,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY c.department DESC, c.number DESC, c.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesNumberAsc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_number)::text IS NULL
    OR (c.number, c.id) > (sqlc.narg(cursor_number)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY c.number ASC, c.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesNumberDesc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_number)::text IS NULL
    OR (c.number, c.id) < (sqlc.narg(cursor_number)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY c.number DESC, c.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesTitleAsc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_title)::text IS NULL
    OR (c.title, c.id) > (sqlc.narg(cursor_title)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY c.title ASC, c.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesTitleDesc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_title)::text IS NULL
    OR (c.title, c.id) < (sqlc.narg(cursor_title)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY c.title DESC, c.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesCreatedAtAsc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (c.created_at, c.id) > (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY c.created_at ASC, c.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListCoursesCreatedAtDesc :many
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE
  (sqlc.narg(school_id)::uuid IS NULL OR c.school_id = sqlc.narg(school_id)::uuid)
  AND (sqlc.narg(department)::text IS NULL OR UPPER(c.department) = UPPER(sqlc.narg(department)::text))
  AND (
    sqlc.narg(q)::text IS NULL
    OR c.title ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
    OR c.department ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR c.number ILIKE sqlc.narg(q)::text ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (c.created_at, c.id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY c.created_at DESC, c.id DESC
LIMIT sqlc.arg(page_limit);

-- name: GetCourse :one
SELECT
  c.id, c.school_id, c.department, c.number, c.title, c.description,
  c.created_at, c.updated_at,
  s.id AS s_id, s.name AS s_name, s.acronym AS s_acronym,
  s.city AS s_city, s.state AS s_state, s.country AS s_country
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE c.id = sqlc.arg(id)::uuid;

-- name: ListCourseSections :many
-- Returns sections with a live member_count via LEFT JOIN (so sections
-- with zero members still appear). Ordered most-recent term first using
-- start_date when present, falling back to section_code. NULLS LAST keeps
-- sections without a known start_date at the bottom.
SELECT
  cs.id, cs.term, cs.section_code, cs.instructor_name, cs.start_date,
  COUNT(cm.user_id) AS member_count
FROM course_sections cs
LEFT JOIN course_members cm ON cm.section_id = cs.id
WHERE cs.course_id = sqlc.arg(course_id)::uuid
GROUP BY cs.id
ORDER BY cs.start_date DESC NULLS LAST, cs.section_code ASC NULLS LAST, cs.id ASC;
