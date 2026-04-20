-- Dashboard queries (ASK-155). Each section of GET /api/me/dashboard
-- has 1-2 dedicated queries; the service orchestrates the ~10 queries
-- in parallel-equivalent fan-out and assembles the response.

-- =============================================================================
-- Current-term resolution waterfall (3 attempts).
-- =============================================================================

-- name: ResolveCurrentTermActive :one
-- Step 1 of the current-term waterfall: find the term whose section
-- window contains today. If multiple terms overlap (rare), use the
-- one with the latest end_date.
SELECT cs.term
FROM course_members cm
JOIN course_sections cs ON cs.id = cm.section_id
WHERE cm.user_id = sqlc.arg(viewer_id)
  AND cs.start_date <= CURRENT_DATE
  AND cs.end_date   >= CURRENT_DATE
GROUP BY cs.term
ORDER BY MAX(cs.end_date) DESC
LIMIT 1;

-- name: ResolveCurrentTermLastEnded :one
-- Step 2 of the waterfall: nothing is active right now (e.g.,
-- between semesters). Pick the term whose latest end_date is
-- closest to today but in the past.
SELECT cs.term
FROM course_members cm
JOIN course_sections cs ON cs.id = cm.section_id
WHERE cm.user_id = sqlc.arg(viewer_id)
  AND cs.end_date IS NOT NULL
  AND cs.end_date <  CURRENT_DATE
ORDER BY cs.end_date DESC
LIMIT 1;

-- name: ResolveCurrentTermLexLatest :one
-- Step 3 of the waterfall: no sections have dates at all (e.g.,
-- newly seeded data). Fall back to lexicographic order, which
-- matches the convention "Spring 2026" > "Fall 2025" alphabetically.
SELECT cs.term
FROM course_members cm
JOIN course_sections cs ON cs.id = cm.section_id
WHERE cm.user_id = sqlc.arg(viewer_id)
ORDER BY cs.term DESC
LIMIT 1;

-- =============================================================================
-- Courses section: enrollment list capped at 10 for the resolved term.
-- =============================================================================

-- name: ListEnrolledCoursesForTerm :many
-- Returns the viewer's enrollments for the resolved current term,
-- capped at 10. The dashboard cares about course-level info plus
-- the viewer's role in that section (TAs, instructors render
-- differently in the UI). joined_at DESC, id DESC for a stable
-- order across calls.
SELECT
  c.id          AS course_id,
  c.department  AS course_department,
  c.number      AS course_number,
  c.title       AS course_title,
  cm.role       AS member_role,
  cs.term       AS section_term
FROM course_members cm
JOIN course_sections cs ON cs.id = cm.section_id
JOIN courses c          ON c.id = cs.course_id
WHERE cm.user_id = sqlc.arg(viewer_id)
  AND cs.term    = sqlc.arg(term)
ORDER BY cm.joined_at DESC, c.id DESC
LIMIT sqlc.arg(page_limit);

-- =============================================================================
-- Study-guides section: count + recent (5).
-- =============================================================================

-- name: CountUserStudyGuides :one
-- Number of non-deleted study guides the viewer created. Filters
-- deleted_at IS NULL because the home page should never surface a
-- count that includes recently-deleted guides.
SELECT COUNT(*)::bigint AS created_count
FROM study_guides
WHERE creator_id = sqlc.arg(viewer_id)
  AND deleted_at IS NULL;

-- name: ListRecentUserStudyGuides :many
-- 5 most recently updated guides the viewer created. Joins courses
-- so the UI can render "<dept> <num> -- <title>" inline. Sorted by
-- updated_at DESC for the home-page "recently updated" affordance.
SELECT
  sg.id           AS study_guide_id,
  sg.title        AS study_guide_title,
  sg.updated_at   AS updated_at,
  c.department    AS course_department,
  c.number        AS course_number
FROM study_guides sg
JOIN courses c ON c.id = sg.course_id
WHERE sg.creator_id = sqlc.arg(viewer_id)
  AND sg.deleted_at IS NULL
ORDER BY sg.updated_at DESC, sg.id DESC
LIMIT sqlc.arg(page_limit);

-- =============================================================================
-- Practice section: aggregate stats + answer count + recent sessions.
-- =============================================================================

-- name: GetUserPracticeStats :one
-- One-shot aggregate over completed sessions: count + sum(correct)
-- + sum(total). The accuracy percentage is computed in Go as
-- ROUND(100 * total_correct / NULLIF(total_questions, 0)) so we
-- can avoid a SQL CASE/COALESCE that would shadow the divide-by-
-- zero behavior. COALESCE on the SUMs returns 0 (not NULL) when
-- the user has no completed sessions -- the spec requires zeros
-- in that case rather than nulls.
SELECT
  COUNT(*) FILTER (WHERE completed_at IS NOT NULL)::bigint                                  AS sessions_completed,
  COALESCE(SUM(correct_answers) FILTER (WHERE completed_at IS NOT NULL), 0)::bigint         AS total_correct,
  COALESCE(SUM(total_questions) FILTER (WHERE completed_at IS NOT NULL), 0)::bigint         AS total_questions
FROM practice_sessions
WHERE user_id = sqlc.arg(viewer_id);

-- name: CountUserAnsweredQuestions :one
-- Per-spec: total_questions_answered is the number of submitted
-- answers across completed sessions, NOT the snapshot total from
-- practice_sessions.total_questions. Computed via a join on
-- practice_answers since that's where actual submissions live.
SELECT COUNT(*)::bigint AS answered_count
FROM practice_answers pa
JOIN practice_sessions ps ON ps.id = pa.session_id
WHERE ps.user_id = sqlc.arg(viewer_id)
  AND ps.completed_at IS NOT NULL;

-- name: ListRecentUserSessions :many
-- 5 most recently completed sessions. Joins quiz + study_guide so
-- each row carries the labels the home page needs (no follow-up
-- GETs). score_percentage is computed in Go from total_questions
-- + correct_answers (consistent with the overall_accuracy formula).
SELECT
  ps.id              AS session_id,
  ps.completed_at    AS completed_at,
  ps.correct_answers AS correct_answers,
  ps.total_questions AS total_questions,
  q.title            AS quiz_title,
  sg.title           AS study_guide_title
FROM practice_sessions ps
JOIN quizzes q       ON q.id = ps.quiz_id
JOIN study_guides sg ON sg.id = q.study_guide_id
WHERE ps.user_id      = sqlc.arg(viewer_id)
  AND ps.completed_at IS NOT NULL
ORDER BY ps.completed_at DESC, ps.id DESC
LIMIT sqlc.arg(page_limit);

-- =============================================================================
-- Files section: aggregate stats + recent files.
-- =============================================================================

-- name: GetUserFileStats :one
-- Total count + total bytes of complete (non-deletion-lifecycle)
-- files the viewer owns. Filters status='complete' because pending
-- and failed uploads aren't user-facing storage.
SELECT
  COUNT(*)::bigint                  AS total_count,
  COALESCE(SUM(size), 0)::bigint    AS total_size
FROM files
WHERE user_id           = sqlc.arg(viewer_id)
  AND deletion_status   IS NULL
  AND status            = 'complete';

-- name: ListRecentUserFiles :many
-- 5 most recently updated complete files the viewer owns.
SELECT
  f.id         AS file_id,
  f.name       AS file_name,
  f.mime_type  AS file_mime_type,
  f.updated_at AS updated_at
FROM files f
WHERE f.user_id           = sqlc.arg(viewer_id)
  AND f.deletion_status   IS NULL
  AND f.status            = 'complete'
ORDER BY f.updated_at DESC, f.id DESC
LIMIT sqlc.arg(page_limit);
