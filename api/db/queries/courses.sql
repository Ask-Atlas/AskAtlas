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

-- name: ListSectionsForCourse :many
-- Dedicated sections endpoint for ASK-127 -- distinct from the
-- inline ListCourseSections (used by GetCourse) because:
--   * Optional exact-match term filter via sqlc.narg.
--   * Returns course_id and created_at (the inline payload omits
--     them because the parent course already carries the id and
--     created_at is irrelevant to the inline render).
--   * Different ORDER BY: term DESC, section_code ASC -- sorts
--     terms in DESCENDING LEXICOGRAPHIC order per the ASK-127
--     spec, with section codes ascending within a term. NOT
--     chronological: "Spring 2026" sorts before "Fall 2026"
--     because S<F false but in DESC order alphabetic decides;
--     "Summer 2025" sorts before "Spring 2025" alphabetically
--     (Su>Sp). Acceptable per the spec. The inline
--     ListCourseSections query sorts by start_date DESC for
--     the chronological course-detail UI instead.
--
-- LEFT JOIN keeps zero-member sections in the result set; COUNT
-- is wrapped in a GROUP BY cs.id so member_count is per-section,
-- not a global aggregate.
SELECT
  cs.id, cs.course_id, cs.term, cs.section_code, cs.instructor_name,
  cs.created_at,
  COUNT(cm.user_id) AS member_count
FROM course_sections cs
LEFT JOIN course_members cm ON cm.section_id = cs.id
WHERE cs.course_id = sqlc.arg(course_id)::uuid
  AND (sqlc.narg(term)::text IS NULL OR cs.term = sqlc.narg(term)::text)
GROUP BY cs.id
ORDER BY cs.term DESC, cs.section_code ASC NULLS LAST, cs.id ASC;

-- name: CourseExists :one
-- Single-row existence probe used by join/leave to disambiguate the
-- "Course not found" 404 from the "Section not found" 404 the spec
-- requires (see ASK-132 / ASK-138).
SELECT EXISTS (
  SELECT 1 FROM courses WHERE id = sqlc.arg(id)::uuid
) AS exists;

-- name: SectionInCourseExists :one
-- Verifies the section exists AND belongs to the supplied course. A
-- section UUID that targets a different course is treated as not found
-- to avoid leaking the existence of unrelated sections via the URL path.
SELECT EXISTS (
  SELECT 1 FROM course_sections
  WHERE id = sqlc.arg(section_id)::uuid
    AND course_id = sqlc.arg(course_id)::uuid
) AS exists;

-- name: JoinSection :one
-- Adds the user to the section as a 'student'. ON CONFLICT DO NOTHING
-- keeps duplicate joins atomic (no PK violation surfacing) and concurrency
-- safe; the service layer treats an empty result as the 409 "Already a
-- member of this section" case.
INSERT INTO course_members (user_id, section_id, role)
VALUES (sqlc.arg(user_id)::uuid, sqlc.arg(section_id)::uuid, 'student')
ON CONFLICT (user_id, section_id) DO NOTHING
RETURNING user_id, section_id, role, joined_at;

-- name: LeaveSection :one
-- Hard-deletes the membership row. RETURNING lets the service detect a
-- no-op delete (sql.ErrNoRows) and map it to the 404 "Not a member of
-- this section" response.
DELETE FROM course_members
WHERE user_id = sqlc.arg(user_id)::uuid
  AND section_id = sqlc.arg(section_id)::uuid
RETURNING user_id;

-- name: ListMyEnrollments :many
-- Returns every section a user is enrolled in with the compact
-- course + school payload the dashboard renders. Optional term + role
-- filters use sqlc.narg so the WHERE branch short-circuits when
-- they're absent. Sort is fixed: most-recent term first (lexicographic
-- "Spring 2026" > "Fall 2025" is acceptable for MVP per the spec),
-- then department + number for stable in-term ordering.
SELECT
  cs.id          AS section_id,
  cs.term        AS section_term,
  cs.section_code AS section_section_code,
  cs.instructor_name AS section_instructor_name,
  c.id           AS course_id,
  c.department   AS course_department,
  c.number       AS course_number,
  c.title        AS course_title,
  s.id           AS school_id,
  s.acronym      AS school_acronym,
  cm.role        AS member_role,
  cm.joined_at   AS member_joined_at
FROM course_members cm
JOIN course_sections cs ON cs.id = cm.section_id
JOIN courses c          ON c.id = cs.course_id
JOIN schools s          ON s.id = c.school_id
WHERE cm.user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(term)::text IS NULL OR cs.term = sqlc.narg(term)::text)
  AND (sqlc.narg(role)::course_role IS NULL OR cm.role = sqlc.narg(role)::course_role)
ORDER BY cs.term DESC, c.department ASC, c.number ASC;

-- name: ListSectionMembers :many
-- Returns the section roster joined against users for first/last name.
-- Privacy floor: SELECT lists ONLY the five fields exposed in the
-- SectionMemberResponse schema (user_id, first_name, last_name, role,
-- joined_at). DO NOT add email, clerk_id, or any other user column to
-- this list -- the endpoint is reachable by any authenticated user.
--
-- Soft-deleted users (users.deleted_at IS NOT NULL) are excluded -- the
-- codebase's soft-delete convention is enforced by the partial indexes
-- idx_users_deleted_at and idx_users_active_email. A user's soft-delete
-- is the signal that they want to disappear from the product, so they
-- must not surface in a public-by-design roster. The cursor still
-- advances past them in the (joined_at, user_id) keyset, so removing
-- them mid-iteration just shrinks the page rather than skipping live
-- members.
--
-- Optional role filter via sqlc.narg short-circuits when absent. Keyset
-- pagination on (joined_at, user_id) -- joined_at alone isn't unique
-- (multiple users can join in the same second on a busy section), so
-- user_id is the tiebreaker that keeps the keyset a strict total order.
SELECT
  cm.user_id,
  u.first_name,
  u.last_name,
  cm.role,
  cm.joined_at
FROM course_members cm
JOIN users u ON u.id = cm.user_id
WHERE cm.section_id = sqlc.arg(section_id)::uuid
  AND u.deleted_at IS NULL
  AND (sqlc.narg(role)::course_role IS NULL OR cm.role = sqlc.narg(role)::course_role)
  AND (
    sqlc.narg(cursor_joined_at)::timestamptz IS NULL
    OR (cm.joined_at, cm.user_id) > (
      sqlc.narg(cursor_joined_at)::timestamptz,
      sqlc.narg(cursor_user_id)::uuid
    )
  )
ORDER BY cm.joined_at ASC, cm.user_id ASC
LIMIT sqlc.arg(page_limit);

-- name: GetMembership :one
-- Single-row membership lookup powering the per-section
-- enrolled/not-enrolled probe (ASK-148). Returns sql.ErrNoRows when the
-- viewer is not a member; the service translates that into the 200
-- {enrolled:false} response, NOT a 404.
SELECT role, joined_at
FROM course_members
WHERE user_id = sqlc.arg(user_id)::uuid
  AND section_id = sqlc.arg(section_id)::uuid;
