-- name: UpsertClerkUser :one
INSERT INTO users (
    clerk_id,
    email,
    first_name,
    last_name,
    middle_name,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, COALESCE(sqlc.arg(metadata), '{}'::jsonb)
)
ON CONFLICT (clerk_id) DO UPDATE 
SET
    email = EXCLUDED.email,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    middle_name = EXCLUDED.middle_name,
    metadata = EXCLUDED.metadata,
    updated_at = NOW(),
    deleted_at = NULL
RETURNING *;

-- name: SoftDeleteUserByClerkID :execrows
UPDATE users
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE
    clerk_id = $1
    AND deleted_at IS NULL;