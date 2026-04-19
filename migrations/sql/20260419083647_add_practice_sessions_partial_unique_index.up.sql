-- ASK-128: enforce at most one INCOMPLETE practice session per (user, quiz).
--
-- Purpose: race protection for POST /api/quizzes/{quiz_id}/sessions.
-- Without this index, two simultaneous start-session requests from the same
-- user against the same quiz can both observe "no incomplete session exists"
-- and both insert a new row -- violating the resume invariant (the user is
-- supposed to see ONE incomplete session at a time per quiz).
--
-- The service layer pairs this index with INSERT ... ON CONFLICT DO NOTHING
-- RETURNING. On conflict, the second request gets back zero rows from the
-- INSERT, re-fetches the existing incomplete session, and returns 200 (resume)
-- instead of 201 (created). The winner of the race never has to retry.
--
-- Partial scope: the index ONLY covers rows with completed_at IS NULL. A
-- user CAN have any number of completed sessions per quiz (retakes are
-- allowed -- AC5). Without the partial WHERE clause, retakes would all
-- collide on the unique constraint.
--
-- Existing query patterns:
--   * The "find my incomplete session" lookup (FindIncompleteSession) reads
--     WHERE user_id=? AND quiz_id=? AND completed_at IS NULL -- this index
--     also covers that read (it's a covering index for the lookup keys).
--   * The "list my sessions on a quiz" lookup (ASK-149, future) is served
--     by idx_practice_sessions_quiz_id; this new index does not affect it.
CREATE UNIQUE INDEX idx_practice_sessions_user_quiz_incomplete
  ON practice_sessions(user_id, quiz_id)
  WHERE completed_at IS NULL;
