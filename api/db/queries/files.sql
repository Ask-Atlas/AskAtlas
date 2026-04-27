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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_mime_type)::text IS NULL
    OR (f.mime_type, f.id) > (sqlc.narg(cursor_mime_type)::text, sqlc.narg(cursor_id)::uuid)
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
  AND (sqlc.narg(mime_type)::text IS NULL OR f.mime_type = sqlc.narg(mime_type)::text)
  AND (sqlc.narg(min_size)::bigint IS NULL OR f.size >= sqlc.narg(min_size)::bigint)
  AND (sqlc.narg(max_size)::bigint IS NULL OR f.size <= sqlc.narg(max_size)::bigint)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR f.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz   IS NULL OR f.created_at <  sqlc.narg(created_to)::timestamptz)
  AND (sqlc.narg(updated_from)::timestamptz IS NULL OR f.updated_at >= sqlc.narg(updated_from)::timestamptz)
  AND (sqlc.narg(updated_to)::timestamptz   IS NULL OR f.updated_at <  sqlc.narg(updated_to)::timestamptz)
  AND (sqlc.narg(q)::text IS NULL OR f.name ILIKE '%' || sqlc.narg(q)::text || '%' ESCAPE '\')
  AND (
    sqlc.narg(cursor_mime_type)::text IS NULL
    OR (f.mime_type, f.id) < (sqlc.narg(cursor_mime_type)::text, sqlc.narg(cursor_id)::uuid)
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

-- name: InsertFileView :exec
-- Append-only analytics row for POST /api/files/{file_id}/view (ASK-134).
-- Each call inserts a fresh row -- no dedup, no upsert. The viewed_at
-- column defaults to now() so the wall-clock stamp lives at the DB
-- layer, not the client. id defaults to gen_random_uuid().
--
-- Existence of the parent file is gated by the service via
-- GetFileForUpdate before this call -- this query trusts inputs.
INSERT INTO file_views (file_id, user_id)
VALUES (sqlc.arg(file_id)::uuid, sqlc.arg(user_id)::uuid);

-- name: UpsertFileLastViewed :exec
-- Per-(user, file) most-recent-view timestamp for POST /api/files/{file_id}/view
-- (ASK-134). Powers the recents sidebar (ASK-145). The PK is
-- (user_id, file_id) so a repeat view by the same user is a write to
-- the same row -- viewed_at gets bumped to now(). On the first view
-- the INSERT path runs; on every subsequent view the ON CONFLICT
-- branch fires.
INSERT INTO file_last_viewed (user_id, file_id, viewed_at)
VALUES (sqlc.arg(user_id)::uuid, sqlc.arg(file_id)::uuid, NOW())
ON CONFLICT (user_id, file_id) DO UPDATE
SET viewed_at = NOW();

-- name: UpdateFileStatus :exec
UPDATE files
SET
    status     = sqlc.arg(status)::upload_status,
    updated_at = NOW()
WHERE id = sqlc.arg(file_id)::uuid
  AND user_id = sqlc.arg(owner_id)::uuid;

-- name: GetFileForExtraction :one
-- Worker-context probe used by the ASK-220 extract worker. Returns the
-- bare minimum the worker needs (s3_key, mime_type, processing_status)
-- so it can decide whether to skip (idempotency: already past
-- 'extracting'), download, parse, or fail. No grant check -- the worker
-- runs as a trusted QStash callback whose authority is the signing-key
-- middleware, not the caller's user_id. Soft-deleted files are filtered
-- so a delete-during-extraction race produces sql.ErrNoRows -> 404 at
-- the handler instead of zombie extraction work.
SELECT id, user_id, s3_key, mime_type, processing_status
FROM files
WHERE id = sqlc.arg(file_id)::uuid
  AND deletion_status IS NULL;

-- name: SetFileProcessingStatus :exec
-- Advances files.processing_status to the next pipeline state. Single
-- writer pattern -- the worker owns the column for the duration of the
-- job; if a parallel run lands here the second one stays idempotent
-- because the worker rechecks via GetFileForExtraction first.
-- Clears status_error on every transition so a retry that succeeds
-- doesn't leave a stale failure message lying around.
UPDATE files
SET
    processing_status = sqlc.arg(processing_status)::processing_status,
    status_error      = NULL,
    updated_at        = NOW()
WHERE id = sqlc.arg(file_id)::uuid;

-- name: MarkFileProcessingFailed :exec
-- Terminal failure path. Sets processing_status='failed' and records
-- the error so the frontend (ASK-222) can surface it. Kept separate
-- from SetFileProcessingStatus because the success path explicitly
-- clears status_error -- collapsing both into one query would force a
-- nullable narg parameter that's easy to misuse.
UPDATE files
SET
    processing_status = 'failed',
    status_error      = sqlc.arg(status_error)::text,
    updated_at        = NOW()
WHERE id = sqlc.arg(file_id)::uuid;

-- name: UpsertExtractedText :exec
-- Stores the extracted plaintext + per-page char-offset array for the
-- chunk+embed worker (ASK-221) to consume. ON CONFLICT DO UPDATE makes
-- the extract worker idempotent on retry: a second successful run
-- silently overwrites the previous text rather than colliding on the
-- file_id PK. ASK-221 is responsible for deleting the row once chunks
-- are written.
INSERT INTO files_extracted_text (file_id, text, page_offsets)
VALUES (
    sqlc.arg(file_id)::uuid,
    sqlc.arg(text)::text,
    sqlc.narg(page_offsets)::integer[]
)
ON CONFLICT (file_id) DO UPDATE SET
    text         = EXCLUDED.text,
    page_offsets = EXCLUDED.page_offsets,
    created_at   = NOW();

-- name: GetExtractedText :one
-- Read-side query for the ASK-221 chunk+embed worker. Returns the
-- extracted plaintext + per-page offsets that ASK-220's extract
-- worker landed in files_extracted_text. The chunk+embed worker
-- consumes this row, splits the text, embeds, persists chunks, and
-- (eventually) deletes this row via DeleteExtractedText. sql.ErrNoRows
-- maps to apperrors.ErrNotFound at the adapter, which the worker
-- treats as terminal-success (the file was deleted between extract
-- and chunk-embed; nothing to do).
SELECT file_id, text, page_offsets
FROM files_extracted_text
WHERE file_id = sqlc.arg(file_id)::uuid;

-- name: DeleteExtractedText :exec
-- Drop the transient extracted-text row after ASK-221 has persisted
-- chunks. The chunks themselves are the canonical store; the
-- extracted-text table is a pipeline handoff only. No FK references
-- this row, so DELETE is unconditionally safe; idempotent if nothing
-- matches.
DELETE FROM files_extracted_text
WHERE file_id = sqlc.arg(file_id)::uuid;

-- name: GetFileForUpdate :one
-- Existence + state probe used by PATCH /api/files/{file_id} (ASK-113).
-- Returns the row's user_id and current status so the service can:
--   * 404 when the row is missing or in any deletion state.
--   * 403 when the caller is not the owner (returned row's user_id mismatch).
--   * Validate status transitions (only pending -> complete / failed allowed).
-- Soft-deleted files are filtered out here so they always map to 404,
-- regardless of caller -- matching the spec's "Resource not found" rule.
SELECT id, user_id, status
FROM files
WHERE id = sqlc.arg(file_id)::uuid
  AND deletion_status IS NULL;

-- name: PatchFile :one
-- Partial update for ASK-113. Each updatable column uses
-- COALESCE(narg, current) so a nil arg means "leave alone" and a
-- non-nil arg means "replace". The CTE returns the post-update row
-- joined with file_favorites + file_last_viewed so the handler can
-- emit a complete FileResponse without a follow-up SELECT.
--
-- The service is responsible for:
--   * 404 / 403 gating (via GetFileForUpdate before this call).
--   * The at-least-one-field rule (an empty body is a 400 before SQL).
--   * Status transition validation (only pending -> complete / failed).
-- Defense-in-depth: the WHERE clause re-asserts owner + non-deleted so
-- a concurrent DELETE between GetFileForUpdate and this UPDATE yields
-- sql.ErrNoRows -> 404 instead of a phantom write.
WITH updated AS (
  UPDATE files
  SET
    name       = COALESCE(sqlc.narg(name)::text,                  name),
    status     = COALESCE(sqlc.narg(status)::upload_status,       status),
    updated_at = NOW()
  WHERE id = sqlc.arg(file_id)::uuid
    AND user_id = sqlc.arg(owner_id)::uuid
    AND deletion_status IS NULL
  RETURNING id, user_id, name, size, mime_type, status, created_at, updated_at
)
SELECT
  u.id, u.user_id, u.name, u.size, u.mime_type, u.status, u.created_at, u.updated_at,
  fav.created_at AS favorited_at,
  lv.viewed_at   AS last_viewed_at
FROM updated u
LEFT JOIN file_favorites  fav ON fav.user_id = sqlc.arg(viewer_id)::uuid AND fav.file_id = u.id
LEFT JOIN file_last_viewed lv ON lv.user_id = sqlc.arg(viewer_id)::uuid AND lv.file_id = u.id;

-- name: InsertFileGrant :one
-- Inserts a new file_grants row for POST /api/files/{file_id}/grants
-- (ASK-122). Plain INSERT -- no ON CONFLICT DO UPDATE because the
-- spec requires returning 409 Conflict on a duplicate (not silently
-- updating granted_by). A unique-key violation (sqlstate 23505)
-- propagates up as a pgx PgError; the service translates it to
-- apperrors.ErrConflict so the handler emits a 409.
INSERT INTO file_grants (file_id, grantee_type, grantee_id, permission, granted_by)
VALUES (
    sqlc.arg(file_id)::uuid,
    sqlc.arg(grantee_type)::grantee_type,
    sqlc.arg(grantee_id)::uuid,
    sqlc.arg(permission)::permission,
    sqlc.arg(granted_by)::uuid
)
RETURNING *;

-- name: RevokeFileGrant :execrows
-- Deletes a file grant matching the exact composite key for DELETE
-- /api/files/{file_id}/grants (ASK-125). Returns the rows-affected
-- count so the service can distinguish "grant exists and was
-- deleted" (1 row -> 204) from "no matching grant" (0 rows -> 404).
-- The spec requires 404 when the grant is missing -- this replaces
-- the previous idempotent no-op behavior.
DELETE FROM file_grants
WHERE file_id = sqlc.arg(file_id)::uuid
  AND grantee_type = sqlc.arg(grantee_type)::grantee_type
  AND grantee_id = sqlc.arg(grantee_id)::uuid
  AND permission = sqlc.arg(permission)::permission;

-- name: CheckUserExists :one
-- Grantee-existence probe for ASK-122 when grantee_type='user'.
-- Returns sql.ErrNoRows when the referenced user does not exist;
-- the service maps this to a 400 VALIDATION_ERROR ("no user with
-- this ID") rather than 404. The public sentinel UUID
-- 00000000-0000-0000-0000-000000000000 is handled in the service
-- layer (skipped before this query ever runs).
SELECT 1
FROM users
WHERE id = sqlc.arg(user_id)::uuid;

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