-- Queries for the batch `POST /api/refs/resolve` endpoint (ASK-208).
-- Each query returns the compact summary shape the frontend uses to
-- render a ref card. Entities missing from the result (deleted,
-- invisible to the viewer, or nonexistent) are absent from the rows;
-- the service fills nulls into the response map.

-- name: ListStudyGuideRefSummaries :many
-- Compact summary for `::sg{id}` refs. Grants-gated (ASK-211): a ref
-- hydrates only when the viewer can see the guide under the same
-- visibility rules as GetStudyGuideDetail -- public, creator, direct
-- user grant, or course grant via enrollment. Guides the viewer can't
-- see simply don't appear in the result set and surface as null refs
-- on the wire (the service layer handles the map lookup). The
-- quiz_count subquery excludes soft-deleted quizzes so a ref to a
-- guide whose only quiz was deleted shows 0.
SELECT
  sg.id,
  sg.title,
  c.department AS course_department,
  c.number     AS course_number,
  (
    SELECT COUNT(*)::int
    FROM quizzes q
    WHERE q.study_guide_id = sg.id
      AND q.deleted_at IS NULL
  )::int AS quiz_count,
  EXISTS (
    SELECT 1
    FROM study_guide_recommendations sgr
    WHERE sgr.study_guide_id = sg.id
  ) AS is_recommended
FROM study_guides sg
JOIN courses c ON c.id = sg.course_id
WHERE sg.id = ANY(sqlc.arg(ids)::uuid[])
  AND sg.deleted_at IS NULL
  AND (
    sg.visibility = 'public'
    OR sg.creator_id = sqlc.arg(viewer_id)::uuid
    OR EXISTS (
      SELECT 1 FROM study_guide_grants g
      WHERE g.study_guide_id = sg.id
        AND g.permission IN ('view', 'share', 'delete')
        AND (
          (g.grantee_type = 'user' AND g.grantee_id = sqlc.arg(viewer_id)::uuid)
          OR (g.grantee_type = 'course' AND EXISTS (
            SELECT 1 FROM course_sections cs
            JOIN course_members cm ON cm.section_id = cs.id
            WHERE cs.course_id = g.grantee_id
              AND cm.user_id = sqlc.arg(viewer_id)::uuid
          ))
        )
    )
  );

-- name: ListQuizRefSummaries :many
-- Compact summary for `::quiz{id}` refs. Same "live parent guide +
-- live creator + not soft-deleted" filter as GetQuizDetail so a
-- hydration that races with a cascade-delete doesn't render an
-- orphaned quiz.
SELECT
  q.id,
  q.title,
  (
    SELECT COUNT(*)::int
    FROM quiz_questions qq
    WHERE qq.quiz_id = q.id
  )::int AS question_count,
  u.first_name AS creator_first_name,
  u.last_name  AS creator_last_name
FROM quizzes q
JOIN users u ON u.id = q.creator_id
JOIN study_guides sg ON sg.id = q.study_guide_id
WHERE q.id = ANY(sqlc.arg(ids)::uuid[])
  AND q.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND sg.deleted_at IS NULL;

-- name: ListFileRefSummaries :many
-- Compact summary for `::file{id}` refs. Grants-gated: owner + direct
-- user grant + public sentinel. Mirrors the GetFileIfViewable
-- visibility branches but fans out over an array of file IDs; course
-- + study_guide grants are intentionally omitted here to match the
-- GetFile handler convention (callers don't resolve viewer course /
-- study-guide IDs upstream yet). Files the viewer can't see simply
-- don't appear in the result set and surface as null refs.
SELECT
  f.id,
  f.name,
  f.size,
  f.mime_type,
  f.status
FROM files f
WHERE f.id = ANY(sqlc.arg(ids)::uuid[])
  AND f.deletion_status IS NULL
  AND (
    f.user_id = sqlc.arg(viewer_id)::uuid
    OR EXISTS (
      SELECT 1
      FROM file_grants g
      WHERE g.file_id = f.id
        AND g.permission IN ('view', 'share', 'delete')
        AND g.grantee_type = 'user'
        AND g.grantee_id = sqlc.arg(viewer_id)::uuid
    )
    OR EXISTS (
      SELECT 1
      FROM file_grants g
      WHERE g.file_id = f.id
        AND g.permission IN ('view', 'share', 'delete')
        AND g.grantee_type = 'user'
        AND g.grantee_id = '00000000-0000-0000-0000-000000000000'
    )
  );

-- name: ListCourseRefSummaries :many
-- Compact summary for `::course{id}` refs. Courses are public; the
-- only filter is the id list. School is joined in-line because the
-- card renders department + number + school.
SELECT
  c.id,
  c.department,
  c.number,
  c.title,
  s.name    AS school_name,
  s.acronym AS school_acronym
FROM courses c
JOIN schools s ON s.id = c.school_id
WHERE c.id = ANY(sqlc.arg(ids)::uuid[]);
