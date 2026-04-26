-- Queries for the chunk + embedding store (ASK-219).
--
-- Two consumer surfaces (separate tickets ship them):
--
--   * ASK-221 chunk+embed worker writes rows after extracting +
--     chunking a file. Idempotent via UPSERT on (file_id, chunk_idx)
--     so re-runs are safe.
--   * ASK-223 grounded generation reads top-k chunks via
--     SearchChunksByEmbedding, filtered to the files the user
--     attached to the active study guide.
--
-- This ticket (ASK-219) only ships the schema + queries; the workers
-- and retrieval endpoint land in separate PRs.

-- name: InsertStudyGuideFileChunk :one
-- Single-row insert with UPSERT semantics: re-running the embed
-- worker on the same (file_id, chunk_idx) overwrites the row,
-- which keeps the worker idempotent across retries.
INSERT INTO study_guide_file_chunks (
  file_id,
  chunk_idx,
  text,
  embedding,
  page,
  heading,
  tokens
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (file_id, chunk_idx) DO UPDATE SET
  text       = EXCLUDED.text,
  embedding  = EXCLUDED.embedding,
  page       = EXCLUDED.page,
  heading    = EXCLUDED.heading,
  tokens     = EXCLUDED.tokens
RETURNING *;

-- name: GetChunksByFile :many
-- All chunks for a file in chunk order. Used by the file-detail
-- diagnostic view + by the worker's count-reconciliation step.
SELECT *
FROM study_guide_file_chunks
WHERE file_id = $1
ORDER BY chunk_idx ASC;

-- name: DeleteChunksByFile :exec
-- Manual chunk delete. Cascade from `files` already removes them on
-- file delete; this is for the worker's "re-chunk this file" path
-- where we want to wipe the previous chunking before re-inserting.
DELETE FROM study_guide_file_chunks
WHERE file_id = $1;

-- name: SearchChunksByEmbedding :many
-- Top-k cosine similarity search over a fixed set of files. The
-- caller (ASK-223 grounded generation) passes the list of file ids
-- the user attached to a study guide so retrieval is naturally
-- scoped to that user's content.
--
-- Returns chunks ordered by ascending cosine *distance*, i.e. most
-- similar first. The HNSW index uses `vector_cosine_ops` so this
-- ORDER BY clause is index-backed.
--
-- The `<=>` operator is pgvector's cosine-distance operator. We
-- expose distance back to the caller (`distance` column) so the
-- ranker can filter on a similarity threshold without a second
-- query.
SELECT
  c.id,
  c.file_id,
  c.chunk_idx,
  c.text,
  c.embedding,
  c.page,
  c.heading,
  c.tokens,
  c.created_at,
  (c.embedding <=> $1)::float8 AS distance
FROM study_guide_file_chunks c
WHERE c.file_id = ANY(@file_ids::uuid[])
ORDER BY c.embedding <=> $1
LIMIT @k::int;
