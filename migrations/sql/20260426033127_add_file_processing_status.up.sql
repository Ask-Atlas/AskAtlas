-- ASK-220: file ingestion pipeline state.
--
-- `files.status` (the existing `upload_status` enum) describes the
-- S3 upload lifecycle (pending -> complete/failed). It STAYS that.
-- This migration adds a parallel `processing_status` column for the
-- post-upload extract -> chunk -> embed pipeline that ASK-220 + ASK-221
-- introduce. Two columns instead of one because the two pipelines are
-- independent: a file can be `status=complete, processing_status=failed`
-- (uploaded fine, extraction broke) or `status=failed, processing_status=uploaded`
-- (S3 upload aborted, never entered the pipeline).
--
-- `processing_status` lifecycle:
--   uploaded  -> initial; pipeline has not started.
--   extracting / extracted   -> ASK-220 worker (this ticket).
--   embedding / ready        -> ASK-221 worker.
--   failed                   -> terminal; status_error carries the cause.
--
-- We backfill existing rows to 'uploaded' so they're eligible for a
-- one-shot reprocessing job later (out of scope here). New rows get the
-- column default.

CREATE TYPE processing_status AS ENUM (
    'uploaded',
    'extracting',
    'extracted',
    'embedding',
    'ready',
    'failed'
);

ALTER TABLE files
    ADD COLUMN processing_status processing_status NOT NULL DEFAULT 'uploaded',
    ADD COLUMN status_error      TEXT;

-- `files_extracted_text` is the handoff between the extract worker
-- (this ticket) and the chunk+embed worker (ASK-221). Lives separately
-- from `files` because the text is large (potentially MBs) and is
-- transient -- ASK-221 deletes the row after chunks are written. PK on
-- file_id makes the upsert + idempotent retries trivial.
--
-- `page_offsets` is a 0-based character-offset-per-page array, used by
-- the chunker to tag each chunk with the page it came from. NULL for
-- sources without page boundaries (text/markdown, text/plain). For PDFs
-- it has one entry per page; entry N is the offset of the first
-- character on page N+1, so chunk-at-offset binary-searches into it.
CREATE TABLE files_extracted_text (
    file_id      UUID         PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
    text         TEXT         NOT NULL,
    page_offsets INTEGER[],
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);
