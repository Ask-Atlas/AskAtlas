-- Restore the original 4-value mime_type enum. This down migration will fail
-- (intentionally) if any files row carries one of the values added in the up
-- migration -- the cast to the enum type cannot succeed for unknown labels.
-- Operators must clean up or convert those rows before rolling back.

ALTER TABLE files
    DROP CONSTRAINT chk_files_mime_type;

CREATE TYPE mime_type AS ENUM (
    'image/jpeg',
    'image/png',
    'image/webp',
    'application/pdf'
);

ALTER TABLE files
    ALTER COLUMN mime_type TYPE mime_type USING mime_type::mime_type;
