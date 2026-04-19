-- Rollback ASK-128: drop the partial unique index on (user_id, quiz_id)
-- WHERE completed_at IS NULL. Reverting this re-opens the duplicate-
-- incomplete-session race -- only roll back if the start-session endpoint
-- is also being rolled back (otherwise concurrent starts will produce
-- duplicate incompletes again).
DROP INDEX IF EXISTS idx_practice_sessions_user_quiz_incomplete;
