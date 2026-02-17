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
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id),
    s3_key     TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    mime_type  mime_type   NOT NULL,
    size       BIGINT      NOT NULL,
    checksum   TEXT,
    status     upload_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- grantee_id is a polymorphic reference resolved by grantee_type:
--   'user'   → users(id)
--   'course' → courses(id)
--   'guide'  → study_guides(id)
-- Public access sentinel: 00000000-0000-0000-0000-000000000000
CREATE TABLE file_grants (
    id           UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id      UUID       NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    grantee_type grantee_type NOT NULL,
    grantee_id   UUID       NOT NULL,
    permission   permission NOT NULL DEFAULT 'view',
    granted_by   UUID       NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (file_id, grantee_type, grantee_id, permission)
);

-- Append-only event log for analytics
CREATE TABLE file_views (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id   UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Most recent view timestamp per user per file (for UX).
CREATE TABLE file_last_viewed (
    file_id   UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (file_id, user_id)
);

CREATE TABLE file_favorites (
    file_id    UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (file_id, user_id)
);


-- Supports "show me all files uploaded by this user" - the primary
-- query for a user's file library.
CREATE INDEX idx_files_user_id
    ON files(user_id);

-- Supports looking up all grants for a given file, e.g. "who has
-- access to this file?" and cascading permission checks on file load.
CREATE INDEX idx_file_grants_file_id
    ON file_grants(file_id);

-- Supports the reverse access-control query: "what files does this
-- grantee (user, course, or guide) have access to?" The composite
-- index on (grantee_type, grantee_id) is required because grantee_id
-- alone is not selective - the same UUID could appear across multiple
-- grantee types.
CREATE INDEX idx_file_grants_grantee
    ON file_grants(grantee_type, grantee_id);

-- Supports analytics queries scoped to a file, e.g. total view count,
-- view history, or a time-series chart of views for a single file.
CREATE INDEX idx_file_views_file_id
    ON file_views(file_id);

-- Supports queries scoped to a user, e.g. "all files this user has
-- viewed"
CREATE INDEX idx_file_views_user_id
    ON file_views(user_id);

-- Composite index supporting "recently viewed files for this user",
-- ordered by recency. The DESC on viewed_at matches the expected sort
-- direction so Postgres can satisfy ORDER BY without a sort step.
CREATE INDEX idx_file_last_viewed_user_viewed
    ON file_last_viewed(user_id, viewed_at DESC);

-- Supports "all files favorited by this user" - the primary query
-- for a user's favorites list.
CREATE INDEX idx_file_favorites_user_id
    ON file_favorites(user_id);

-- Supports "all users who favorited this file", e.g. for displaying
-- a favorite count on a file detail page.
CREATE INDEX idx_file_favorites_file_id
    ON file_favorites(file_id);