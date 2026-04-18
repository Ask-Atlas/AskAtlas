-- name: GetSchool :one
SELECT *
FROM schools
WHERE id = sqlc.arg(id)::uuid;

-- name: ListSchools :many
SELECT *
FROM schools
WHERE (
    sqlc.narg(q)::text IS NULL
    OR acronym ILIKE sqlc.narg(q)::text ESCAPE '\'
    OR name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\'
  )
  AND (
    sqlc.narg(cursor_name)::text IS NULL
    OR (name, id) > (sqlc.narg(cursor_name)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY name ASC, id ASC
LIMIT sqlc.arg(page_limit);
