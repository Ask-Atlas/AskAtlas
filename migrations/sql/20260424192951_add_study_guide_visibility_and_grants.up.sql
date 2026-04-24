-- ASK-207: study-guide visibility + grants.
--
-- Introduces the same grant pattern used by file_grants (ASK-122) to
-- study guides. The viewer can access a guide if the guide is public,
-- they're the creator, or a matching study_guide_grants row exists.

CREATE TYPE study_guide_visibility AS ENUM ('private', 'public');

ALTER TABLE study_guides
  ADD COLUMN visibility study_guide_visibility NOT NULL DEFAULT 'private';

CREATE TABLE study_guide_grants (
  id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  study_guide_id UUID         NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
  grantee_type   grantee_type NOT NULL,
  grantee_id     UUID         NOT NULL,
  permission     permission   NOT NULL DEFAULT 'view',
  granted_by     UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (study_guide_id, grantee_type, grantee_id, permission),
  CHECK (grantee_type IN ('user', 'course'))
);

CREATE INDEX idx_study_guide_grants_grantee_perm_guide
  ON study_guide_grants(grantee_type, grantee_id, permission, study_guide_id);

CREATE INDEX idx_study_guide_grants_sg_id
  ON study_guide_grants(study_guide_id);

-- Backfill: every existing (non-deleted) guide gets a 'course' view
-- grant scoped to its current course_id so current readers don't
-- lose access when the visibility filter flips on. Creators are
-- always able to view regardless of grants -- no backfill needed.
INSERT INTO study_guide_grants (study_guide_id, grantee_type, grantee_id, permission, granted_by)
SELECT sg.id, 'course'::grantee_type, sg.course_id, 'view'::permission, sg.creator_id
FROM study_guides sg
WHERE sg.deleted_at IS NULL;
