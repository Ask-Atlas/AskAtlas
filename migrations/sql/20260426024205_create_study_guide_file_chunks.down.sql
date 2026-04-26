-- ASK-219 rollback. Indexes drop with the table.
--
-- Drop the extension last + only when nothing else uses it. ASK-219
-- is the first user of pgvector so this is safe; future migrations
-- that add other vector columns will need to leave the extension in
-- place.

DROP TABLE IF EXISTS study_guide_file_chunks;

DROP EXTENSION IF EXISTS vector;
