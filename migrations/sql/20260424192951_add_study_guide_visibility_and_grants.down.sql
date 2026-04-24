-- Reverse ASK-207.

DROP INDEX IF EXISTS idx_study_guide_grants_sg_id;
DROP INDEX IF EXISTS idx_study_guide_grants_grantee_perm_guide;
DROP TABLE IF EXISTS study_guide_grants;
ALTER TABLE study_guides DROP COLUMN IF EXISTS visibility;
DROP TYPE IF EXISTS study_guide_visibility;
