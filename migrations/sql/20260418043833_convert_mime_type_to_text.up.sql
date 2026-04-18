-- Convert files.mime_type from the mime_type enum to TEXT with a CHECK
-- constraint. The original enum was too restrictive: PostgreSQL caps enum
-- labels at 63 bytes, but the OOXML mime types we need (.docx, .pptx) are
-- 71 and 75 bytes respectively. TEXT + CHECK lifts that limit while keeping
-- DB-side validation, and future additions only need a CHECK swap (no
-- column rewrite, no ALTER TYPE outside a transaction).

ALTER TABLE files
    ALTER COLUMN mime_type TYPE TEXT USING mime_type::text;

ALTER TABLE files
    ADD CONSTRAINT chk_files_mime_type CHECK (
        mime_type IN (
            'image/jpeg',
            'image/png',
            'image/webp',
            'application/pdf',
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
            'application/vnd.openxmlformats-officedocument.presentationml.presentation',
            'text/plain',
            'application/epub+zip'
        )
    );

DROP TYPE mime_type;
