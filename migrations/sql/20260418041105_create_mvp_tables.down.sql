-- Drop in reverse dependency order so FK references resolve cleanly.

DROP TABLE IF EXISTS course_last_viewed;
DROP TABLE IF EXISTS course_favorites;
DROP TABLE IF EXISTS study_guide_last_viewed;
DROP TABLE IF EXISTS study_guide_favorites;
DROP TABLE IF EXISTS study_guide_files;
DROP TABLE IF EXISTS course_files;
DROP TABLE IF EXISTS study_guide_resources;
DROP TABLE IF EXISTS course_resources;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS practice_answers;
DROP TABLE IF EXISTS practice_session_questions;
DROP TABLE IF EXISTS practice_sessions;
DROP TABLE IF EXISTS quiz_answer_options;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS quizzes;
DROP TABLE IF EXISTS study_guide_recommendations;
DROP TABLE IF EXISTS study_guide_votes;
DROP TABLE IF EXISTS study_guides;
DROP TABLE IF EXISTS course_members;
DROP TABLE IF EXISTS course_sections;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS schools;

DROP TYPE IF EXISTS vote_direction;
DROP TYPE IF EXISTS resource_type;
DROP TYPE IF EXISTS question_type;
DROP TYPE IF EXISTS course_role;
