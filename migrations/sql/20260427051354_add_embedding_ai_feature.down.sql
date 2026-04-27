-- ASK-221 rollback: deliberately a no-op.
--
-- Postgres has no `ALTER TYPE ... DROP VALUE`. Removing the
-- 'embedding' label requires recreating ai_feature without it +
-- updating every column that references the old type, which is risky
-- when ai_usage already holds rows. Since this migration is purely
-- additive and the label has no rows in the rolled-back state, the
-- safe rollback is to leave it in place.
--
-- If we ever need to actually drop the label, the path is: scrub
-- ai_usage rows where feature = 'embedding', recreate the type, then
-- run a column-by-column USING cast. Out of scope here.

SELECT 1; -- explicit no-op so migrate's empty-file guard doesn't bark
