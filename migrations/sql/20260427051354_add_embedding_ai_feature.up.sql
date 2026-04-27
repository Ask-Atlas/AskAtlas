-- ASK-221: extend ai_feature with 'embedding' so the chunk+embed
-- worker's per-file OpenAI embedding spend is captured by the same
-- ai_usage ledger ASK-214 set up. Without this label the worker
-- can't write a row tagged to its surface, and either has to mis-tag
-- as 'other' (loses cost-attribution granularity) or skip the ledger
-- entirely (drops embedding cost from quota + audit reporting).
--
-- ALTER TYPE ... ADD VALUE is non-transactional in older Postgres,
-- but Postgres 12+ allows it inside a transaction as long as the new
-- value is not used in the same statement. golang-migrate wraps each
-- migration in a tx; we're on PG 17 (Neon), so this is safe.

ALTER TYPE ai_feature ADD VALUE IF NOT EXISTS 'embedding';
