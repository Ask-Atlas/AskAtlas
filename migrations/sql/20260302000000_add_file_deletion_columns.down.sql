ALTER TABLE files
  DROP COLUMN IF EXISTS deletion_job_id,
  DROP COLUMN IF EXISTS s3_deleted_at,
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS deletion_status;

DROP TYPE IF EXISTS file_deletion_status;
