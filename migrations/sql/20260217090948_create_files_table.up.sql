CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE grantee_type AS ENUM ('user', 'course', 'study_guide');
CREATE TYPE upload_status AS ENUM ('pending', 'complete', 'failed');
CREATE TYPE mime_type AS ENUM (
  'image/jpeg',
  'image/png',
  'image/webp',
  'application/pdf'
);
CREATE TYPE permission AS ENUM ('view', 'share', 'delete');

CREATE TABLE files (
  id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  s3_key     TEXT          NOT NULL,
  name       TEXT          NOT NULL,
  mime_type  mime_type     NOT NULL,
  size       BIGINT        NOT NULL,
  checksum   TEXT,
  status     upload_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TABLE file_grants (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id      UUID         NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  grantee_type grantee_type NOT NULL,
  grantee_id   UUID         NOT NULL,
  permission   permission   NOT NULL DEFAULT 'view',
  granted_by   UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (file_id, grantee_type, grantee_id, permission)
);

CREATE TABLE file_views (
  id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id   UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  viewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Most recent view timestamp per user per file (for UX).
-- PK order optimized for "list recents for a user".
CREATE TABLE file_last_viewed (
  user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  file_id   UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, file_id)
);

-- User favorites (for UX).
-- PK order optimized for "list favorites for a user".
CREATE TABLE file_favorites (
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  file_id    UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, file_id)
);

-- Supports "show me all files uploaded by this user", sorted by created_at
-- with stable keyset pagination using (created_at, id).
CREATE INDEX idx_files_user_created_id
  ON files(user_id, created_at, id);

-- Supports "my files sorted by last update", with stable keyset pagination
-- using (updated_at, id).
CREATE INDEX idx_files_user_updated_id
  ON files(user_id, updated_at, id);

-- Supports "my files sorted by name" (case-insensitive).
-- Use ORDER BY lower(name), id in queries to match this index.
CREATE INDEX idx_files_user_lower_name_id
  ON files(user_id, lower(name), id);

-- Supports "my files sorted by size", with stable keyset pagination
-- using (size, id).
CREATE INDEX idx_files_user_size_id
  ON files(user_id, size, id);

-- Supports "my files sorted by status", with stable keyset pagination
-- using (status, id).
CREATE INDEX idx_files_user_status_id
  ON files(user_id, status, id);

-- Supports "my files sorted by mime type", with stable keyset pagination
-- using (mime_type, id).
CREATE INDEX idx_files_user_mime_id
  ON files(user_id, mime_type, id);

-- Supports fast filename "contains" search (ILIKE '%term%').
-- Used in both owned and accessible file list views.
CREATE INDEX idx_files_name_trgm
  ON files USING GIN (name gin_trgm_ops);

-- Supports the reverse access-control query:
-- "what files does this grantee (user/course/study_guide) have access to?"
-- and permission checks like: grantee + required_permission -> file_id.
-- Note: enum order is permission hierarchy ('view' < 'share' < 'delete'),
CREATE INDEX idx_file_grants_grantee_perm_file
  ON file_grants(grantee_type, grantee_id, permission, file_id);

-- Supports looking up all grants for a given file:
-- "who has access to this file?"
CREATE INDEX idx_file_grants_file_id
  ON file_grants(file_id);

-- Supports analytics queries scoped to a file, e.g. total view count,
-- view history, or time-series chart of views for a file.
CREATE INDEX idx_file_views_file_id
  ON file_views(file_id);

-- Supports analytics queries scoped to a user, e.g. "all files this user has viewed".
CREATE INDEX idx_file_views_user_id
  ON file_views(user_id);

-- Supports "recently viewed files for this user", ordered by recency,
-- with stable keyset pagination using (viewed_at, file_id).
CREATE INDEX idx_file_last_viewed_user_viewed_file
  ON file_last_viewed(user_id, viewed_at, file_id);

-- Supports "all files favorited by this user", ordered by favorite date,
-- with stable keyset pagination using (created_at, file_id).
CREATE INDEX idx_file_favorites_user_created_file
  ON file_favorites(user_id, created_at, file_id);

-- Supports "all users who favorited this file" (and fast favorite counts).
CREATE INDEX idx_file_favorites_file_id
  ON file_favorites(file_id);