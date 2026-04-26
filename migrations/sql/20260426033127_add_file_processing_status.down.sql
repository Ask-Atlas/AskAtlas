-- ASK-220 rollback. Drop the table first (cascades on file_id), then
-- the columns, then the enum type. Order matters: the enum can only
-- drop after the column that uses it is gone.

DROP TABLE IF EXISTS files_extracted_text;

ALTER TABLE files
    DROP COLUMN IF EXISTS status_error,
    DROP COLUMN IF EXISTS processing_status;

DROP TYPE IF EXISTS processing_status;
