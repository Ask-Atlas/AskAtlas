
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_token VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    middle_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Index for lookups by Clerk token
CREATE INDEX idx_users_clerk_token ON users(clerk_token);

-- Index for filtering active users
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Index for email lookups
CREATE INDEX idx_users_email ON users(email);