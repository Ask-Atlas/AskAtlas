CREATE TYPE file_deletion_status AS ENUM ('pending_deletion', 'deleted');

ALTER TABLE files
  ADD COLUMN deletion_status  file_deletion_status NULL,
  ADD COLUMN deleted_at       TIMESTAMPTZ          NULL,
  ADD COLUMN s3_deleted_at    TIMESTAMPTZ          NULL,
  ADD COLUMN deletion_job_id  TEXT                 NULL;
