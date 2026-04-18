-- name: ListOwnedFilesUpdatedDesc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_updated_at)::timestamptz IS NULL
    OR (f.updated_at, f.id) < (sqlc.narg(cursor_updated_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.updated_at DESC, f.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesUpdatedAsc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_updated_at)::timestamptz IS NULL
    OR (f.updated_at, f.id) > (sqlc.narg(cursor_updated_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.updated_at ASC, f.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesCreatedDesc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (f.created_at, f.id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.created_at DESC, f.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesCreatedAsc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (f.created_at, f.id) > (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.created_at ASC, f.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesNameAsc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_name_lower)::text IS NULL
    OR (lower(f.name), f.id) > (sqlc.narg(cursor_name_lower)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY lower(f.name) ASC, f.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesNameDesc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_name_lower)::text IS NULL
    OR (lower(f.name), f.id) < (sqlc.narg(cursor_name_lower)::text, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY lower(f.name) DESC, f.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesSizeAsc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_size)::bigint IS NULL
    OR (f.size, f.id) > (sqlc.narg(cursor_size)::bigint, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.size ASC, f.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesSizeDesc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_size)::bigint IS NULL
    OR (f.size, f.id) < (sqlc.narg(cursor_size)::bigint, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.size DESC, f.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesStatusAsc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_status)::upload_status IS NULL
    OR (f.status, f.id) > (sqlc.narg(cursor_status)::upload_status, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.status ASC, f.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesStatusDesc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_status)::upload_status IS NULL
    OR (f.status, f.id) < (sqlc.narg(cursor_status)::upload_status, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.status DESC, f.id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesMimeAsc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_mime_type)::mime_type IS NULL
    OR (f.mime_type, f.id) > (sqlc.narg(cursor_mime_type)::mime_type, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.mime_type ASC, f.id ASC
LIMIT sqlc.arg(page_limit);

-- name: ListOwnedFilesMimeDesc :many
SELECT
  f.*,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM files f
LEFT JOIN file_favorites fav
  ON fav.user_id = sqlc.arg(viewer_id) AND fav.file_id = f.id
LEFT JOIN file_last_viewed lv
  ON lv.user_id = sqlc.arg(viewer_id) AND lv.file_id = f.id
WHERE f.user_id = sqlc.arg(owner_id)
  AND f.deletion_status IS NULL
  AND (sqlc.narg(status)::upload_status IS NULL OR f.status = sqlc.narg(status)::upload_status)
  AND (sqlc.narg(mime_type)::mime_type IS NULL OR f.mime_type = sqlc.narg(mime_type)::mime_type)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_mime_type)::mime_type IS NULL
    OR (f.mime_type, f.id) < (sqlc.narg(cursor_mime_type)::mime_type, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY f.mime_type DESC, f.id DESC
LIMIT sqlc.arg(page_limit);

-- name: GetFileByOwner :one
-- Fetches a file only if it belongs to the given user and has not been soft-deleted.
-- Returns sql.ErrNoRows if not found or already in a deletion state.
SELECT id, user_id, s3_key, name, mime_type, size, checksum,
       status, deletion_status, deleted_at, s3_deleted_at, deletion_job_id,
       created_at, updated_at
FROM files
WHERE id = sqlc.arg(file_id)::uuid
  AND user_id = sqlc.arg(owner_id)::uuid
  AND deletion_status IS NULL;

-- name: SoftDeleteFile :execrows
-- Marks a file as pending deletion. Only applies if the file is owned by the caller
-- and has not already entered a deletion state (idempotency-safe).
UPDATE files
SET
    deletion_status = 'pending_deletion',
    deleted_at      = NOW(),
    updated_at      = NOW()
WHERE id      = sqlc.arg(file_id)::uuid
  AND user_id = sqlc.arg(owner_id)::uuid
  AND deletion_status IS NULL;

-- name: SetFileDeletionJobID :exec
-- Records the QStash message ID after publishing the async cleanup job.
UPDATE files
SET
    deletion_job_id = sqlc.arg(job_id),
    updated_at      = NOW()
WHERE id = sqlc.arg(file_id)::uuid
  AND deletion_status = 'pending_deletion';

-- name: MarkFileDeleted :exec
-- Called by the cleanup job handler once S3 deletion is confirmed.
UPDATE files
SET
    deletion_status = 'deleted',
    s3_deleted_at   = NOW(),
    updated_at      = NOW()
WHERE id = sqlc.arg(file_id)::uuid
  AND deletion_status = 'pending_deletion';

-- name: InsertFile :one
INSERT INTO files (id, user_id, s3_key, name, mime_type, size, status)
VALUES ($1, $2, $3, $4, $5, $6, 'pending')
RETURNING *;

-- name: UpdateFileStatus :exec
UPDATE files
SET
    status     = sqlc.arg(status)::upload_status,
    updated_at = NOW()
WHERE id = sqlc.arg(file_id)::uuid
  AND user_id = sqlc.arg(owner_id)::uuid;

-- name: UpdateFile :one
-- Renames a file. Only applies if owned by the caller and not in a deletion state.
-- Returns sql.ErrNoRows when file is not found, not owned, or in deletion.
UPDATE files
SET
    name       = sqlc.arg(name),
    updated_at = NOW()
WHERE id      = sqlc.arg(file_id)::uuid
  AND user_id = sqlc.arg(owner_id)::uuid
  AND deletion_status IS NULL
RETURNING id, user_id, name, size, mime_type, status, created_at, updated_at;

-- name: GetFileIfViewable :one
SELECT f.*
FROM files f
WHERE f.id = sqlc.arg(file_id)
  AND f.deletion_status IS NULL
  AND (
    -- owner always can view
    f.user_id = sqlc.arg(viewer_id)

    -- public sentinel
    OR EXISTS (
      SELECT 1
      FROM file_grants g
      WHERE g.file_id = f.id
        AND g.permission IN ('view', 'share', 'delete')
        AND g.grantee_type = 'user'
        AND g.grantee_id = '00000000-0000-0000-0000-000000000000'
    )

    -- direct user grant
    OR EXISTS (
      SELECT 1
      FROM file_grants g
      WHERE g.file_id = f.id
        AND g.permission IN ('view', 'share', 'delete')
        AND g.grantee_type = 'user'
        AND g.grantee_id = sqlc.arg(viewer_id)
    )

    -- course grant
    OR EXISTS (
      SELECT 1
      FROM file_grants g
      WHERE g.file_id = f.id
        AND g.permission IN ('view', 'share', 'delete')
        AND g.grantee_type = 'course'
        AND g.grantee_id = ANY(sqlc.arg(course_ids)::uuid[])
    )

    -- study_guide grant
    OR EXISTS (
      SELECT 1
      FROM file_grants g
      WHERE g.file_id = f.id
        AND g.permission IN ('view', 'share', 'delete')
        AND g.grantee_type = 'study_guide'
        AND g.grantee_id = ANY(sqlc.arg(study_guide_ids)::uuid[])
    )
  );