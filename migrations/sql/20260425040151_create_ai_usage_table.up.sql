-- ASK-214: AI rate limit + cost ledger.
--
-- Persistent record of every Claude/OpenAI request the API
-- dispatches. Two consumers:
--
--   1. quota service -- count today's spend per (user, feature) and
--      reject with 429 when the daily limit is exceeded.
--   2. billing/audit -- exact token counts (incl. prompt-cache reads)
--      so we can reconcile invoices and answer "why was this user
--      charged so much?".
--
-- The table holds NO content -- no prompts, no responses. Chat
-- transcripts, when we want them, live in a separate ai_messages
-- table (ASK-230) with very different retention + privacy needs.

CREATE TYPE ai_feature AS ENUM (
  'ping',
  'edit',
  'grounded_edit',
  'qa',
  'quiz',
  'ref_suggest',
  'other'
);

CREATE TABLE ai_usage (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  feature            ai_feature  NOT NULL,
  model              TEXT        NOT NULL,
  input_tokens       BIGINT      NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
  output_tokens      BIGINT      NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
  cache_read_tokens  BIGINT      NOT NULL DEFAULT 0 CHECK (cache_read_tokens >= 0),
  cache_write_tokens BIGINT      NOT NULL DEFAULT 0 CHECK (cache_write_tokens >= 0),
  request_id         TEXT        NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- (user_id, feature, created_at desc) supports the dominant query:
-- "how many <feature> rows for this user since UTC midnight?". The
-- created_at descending order matches both the count predicate and
-- any future "show me my last N AI calls" UI.
CREATE INDEX idx_ai_usage_user_feature_created
  ON ai_usage (user_id, feature, created_at DESC);

-- request_id is unique on the wire (uuid v4) but we don't enforce
-- uniqueness in the DB -- the cost-log path can write twice on
-- legitimate retries (e.g. the SDK retried a 5xx). Counting wins
-- over deduplication for billing accuracy.
CREATE INDEX idx_ai_usage_request_id
  ON ai_usage (request_id);
