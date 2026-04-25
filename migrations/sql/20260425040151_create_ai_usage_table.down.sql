-- ASK-214 rollback. Drops the table first so the FK-on-CASCADE chain
-- doesn't surprise us, then the supporting enum.

DROP TABLE IF EXISTS ai_usage;
DROP TYPE  IF EXISTS ai_feature;
