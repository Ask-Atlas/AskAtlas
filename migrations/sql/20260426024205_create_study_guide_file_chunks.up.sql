-- ASK-219: pgvector + study_guide_file_chunks schema.
--
-- Foundation for the grounded-AI epic (retrieval-augmented edit /
-- generate / Q&A). One row per chunk extracted from a file the user
-- attached to a study guide. Chunks carry their own OpenAI
-- text-embedding-3-small (1536-dim) vector for cosine similarity
-- search. Actual extraction + embedding workers ship in ASK-220 +
-- ASK-221; this migration just lays down the schema + index.
--
-- Provider note: we're on Neon, not Supabase. Neon ships pgvector
-- but NOT pgvectorscale (Timescale's diskann index). Plain HNSW
-- from pgvector is sufficient for MVP scale (well under 50M
-- vectors); revisit if recall or latency degrades.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE study_guide_file_chunks (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id     UUID         NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  -- 0-based ordinal of this chunk within the file. Unique with
  -- file_id so re-running the embed worker (ASK-221) is idempotent
  -- via UPSERT on (file_id, chunk_idx).
  chunk_idx   INTEGER      NOT NULL CHECK (chunk_idx >= 0),
  text        TEXT         NOT NULL,
  -- 1536-dim matches OpenAI text-embedding-3-small. If we ever
  -- swap models with a different dim, that's a new column + a
  -- backfill, not an in-place change.
  embedding   vector(1536) NOT NULL,
  -- PDF page (1-based) when the source had one. NULL for plain text
  -- and markdown sources. Surfaced in ASK-224 hover cards.
  page        INTEGER      CHECK (page IS NULL OR page > 0),
  -- Nearest preceding markdown heading text, when the chunker can
  -- find one. Used by the citation hover card and to bias retrieval.
  heading     TEXT,
  -- Token count for the chunker's tokenization (roughly tiktoken).
  -- Logged at retrieval time so the cost ledger (ASK-214) can
  -- attribute embedding spend back to specific files.
  tokens      INTEGER      NOT NULL CHECK (tokens >= 0),
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),

  UNIQUE (file_id, chunk_idx)
);

-- HNSW with cosine ops. Pure pgvector (Neon-compatible). Defaults
-- (m=16, ef_construction=64) are fine until recall or latency
-- starts hurting; tune at that point, not preemptively.
CREATE INDEX idx_study_guide_file_chunks_embedding_hnsw
  ON study_guide_file_chunks USING hnsw (embedding vector_cosine_ops);

-- Supports "all chunks for this file in chunk order" -- used by the
-- file-detail view + by the embed worker when reconciling counts.
CREATE INDEX idx_study_guide_file_chunks_file_idx
  ON study_guide_file_chunks (file_id, chunk_idx);
