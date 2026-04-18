-- =============================================================================
-- Enums
-- =============================================================================

CREATE TYPE course_role AS ENUM ('student', 'instructor', 'ta');
CREATE TYPE question_type AS ENUM ('multiple_choice', 'true_false', 'freeform');
CREATE TYPE resource_type AS ENUM ('link', 'video', 'article', 'pdf');
CREATE TYPE vote_direction AS ENUM ('up', 'down');

-- =============================================================================
-- Schools & Courses
-- =============================================================================

CREATE TABLE schools (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    acronym    VARCHAR(20) NOT NULL,
    domain     TEXT,
    url        TEXT,
    city       TEXT,
    state      TEXT,
    country    VARCHAR(2),
    ipeds_id   VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_schools_acronym ON schools(acronym);
CREATE INDEX idx_schools_name ON schools USING GIN (name gin_trgm_ops);
CREATE UNIQUE INDEX idx_schools_ipeds_id ON schools(ipeds_id) WHERE ipeds_id IS NOT NULL;
CREATE UNIQUE INDEX idx_schools_domain ON schools(domain) WHERE domain IS NOT NULL;

CREATE TABLE courses (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID        NOT NULL REFERENCES schools(id) ON DELETE RESTRICT,
    department  VARCHAR(20) NOT NULL,
    number      VARCHAR(20) NOT NULL,
    title       TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_courses_school_dept_num UNIQUE (school_id, department, number)
);

CREATE INDEX idx_courses_school_id ON courses(school_id);
CREATE INDEX idx_courses_title ON courses USING GIN (title gin_trgm_ops);

CREATE TABLE course_sections (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       UUID        NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    term            VARCHAR(30) NOT NULL,
    section_code    VARCHAR(20),
    instructor_name TEXT,
    start_date      DATE,
    end_date        DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_sections_course_term_code UNIQUE (course_id, term, section_code),
    CONSTRAINT chk_sections_dates CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date)
);

CREATE INDEX idx_course_sections_course_id ON course_sections(course_id);
CREATE INDEX idx_course_sections_dates ON course_sections(start_date, end_date) WHERE start_date IS NOT NULL;
CREATE UNIQUE INDEX uq_sections_one_uncoded ON course_sections(course_id, term) WHERE section_code IS NULL;

CREATE TABLE course_members (
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    section_id UUID        NOT NULL REFERENCES course_sections(id) ON DELETE CASCADE,
    role       course_role NOT NULL DEFAULT 'student',
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, section_id)
);

CREATE INDEX idx_course_members_section_id ON course_members(section_id);

-- =============================================================================
-- Study Guides
-- =============================================================================

CREATE TABLE study_guides (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   UUID        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    creator_id  UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title       TEXT        NOT NULL,
    description TEXT,
    content     TEXT,
    tags        TEXT[]      NOT NULL DEFAULT '{}',
    view_count  INTEGER     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_study_guides_course_id ON study_guides(course_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_study_guides_creator_id ON study_guides(creator_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_study_guides_tags ON study_guides USING GIN (tags) WHERE deleted_at IS NULL;

CREATE TABLE study_guide_votes (
    user_id        UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    study_guide_id UUID           NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    vote           vote_direction NOT NULL,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, study_guide_id)
);

CREATE INDEX idx_study_guide_votes_guide_id ON study_guide_votes(study_guide_id);

CREATE TABLE study_guide_recommendations (
    study_guide_id UUID        NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    recommended_by UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (study_guide_id, recommended_by)
);

CREATE INDEX idx_study_guide_recommendations_user ON study_guide_recommendations(recommended_by);

-- =============================================================================
-- Quizzes
-- =============================================================================

CREATE TABLE quizzes (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    study_guide_id UUID        NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    creator_id     UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title          TEXT        NOT NULL,
    description    TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_quizzes_study_guide_id ON quizzes(study_guide_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_quizzes_creator_id ON quizzes(creator_id) WHERE deleted_at IS NULL;

CREATE TABLE quiz_questions (
    id                 UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id            UUID          NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    type               question_type NOT NULL,
    question_text      TEXT          NOT NULL,
    hint               TEXT,
    feedback_correct   TEXT,
    feedback_incorrect TEXT,
    reference_answer   TEXT,
    is_protected       BOOLEAN       NOT NULL DEFAULT false,
    sort_order         INTEGER       NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_quiz_questions_quiz_id ON quiz_questions(quiz_id);

CREATE TABLE quiz_answer_options (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID    NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    text        TEXT    NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT false,
    sort_order  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_quiz_answer_options_question_id ON quiz_answer_options(question_id);

-- =============================================================================
-- Practice Sessions
-- =============================================================================

CREATE TABLE practice_sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id         UUID        NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    total_questions INTEGER     NOT NULL DEFAULT 0,
    correct_answers INTEGER     NOT NULL DEFAULT 0
);

CREATE INDEX idx_practice_sessions_user_id ON practice_sessions(user_id);
CREATE INDEX idx_practice_sessions_quiz_id ON practice_sessions(quiz_id);
CREATE INDEX idx_practice_sessions_user_completed
    ON practice_sessions(user_id, completed_at DESC NULLS LAST);

-- practice_session_questions uses a surrogate id PK because question_id is
-- nullable (ON DELETE SET NULL from quiz_questions) and PK columns can't be
-- NULL. Uniqueness of (session_id, question_id) is enforced via a partial
-- unique index that excludes the NULLed rows.
CREATE TABLE practice_session_questions (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID    NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
    question_id UUID    REFERENCES quiz_questions(id) ON DELETE SET NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_practice_session_questions_session_question
    ON practice_session_questions(session_id, question_id) WHERE question_id IS NOT NULL;
CREATE INDEX idx_practice_session_questions_session_id ON practice_session_questions(session_id);
CREATE INDEX idx_practice_session_questions_question_id ON practice_session_questions(question_id) WHERE question_id IS NOT NULL;

-- verified=true marks server-validated answers (MCQ/TF). Freeform answers may
-- arrive verified=false until graded.
CREATE TABLE practice_answers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID        NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
    question_id UUID        REFERENCES quiz_questions(id) ON DELETE SET NULL,
    user_answer TEXT,
    is_correct  BOOLEAN,
    verified    BOOLEAN     NOT NULL DEFAULT true,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_practice_answers_session_question UNIQUE (session_id, question_id)
);

CREATE INDEX idx_practice_answers_session_id ON practice_answers(session_id);

-- =============================================================================
-- Resources (external URLs, independent entity)
-- =============================================================================

CREATE TABLE resources (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id  UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title       TEXT          NOT NULL,
    url         TEXT          NOT NULL,
    description TEXT,
    type        resource_type NOT NULL DEFAULT 'link',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_resources_creator_id ON resources(creator_id);
CREATE UNIQUE INDEX idx_resources_creator_url ON resources(creator_id, url);

-- =============================================================================
-- Join tables: resources <-> courses/study_guides
-- =============================================================================

CREATE TABLE course_resources (
    resource_id UUID        NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    course_id   UUID        NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    attached_by UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (resource_id, course_id)
);

CREATE INDEX idx_course_resources_course_id ON course_resources(course_id);

CREATE TABLE study_guide_resources (
    resource_id    UUID        NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    study_guide_id UUID        NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    attached_by    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (resource_id, study_guide_id)
);

CREATE INDEX idx_study_guide_resources_guide_id ON study_guide_resources(study_guide_id);

-- =============================================================================
-- Join tables: files <-> courses/study_guides
-- =============================================================================

CREATE TABLE course_files (
    file_id    UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    course_id  UUID        NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (file_id, course_id)
);

CREATE INDEX idx_course_files_course_id ON course_files(course_id);

CREATE TABLE study_guide_files (
    file_id        UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    study_guide_id UUID        NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (file_id, study_guide_id)
);

CREATE INDEX idx_study_guide_files_guide_id ON study_guide_files(study_guide_id);

-- =============================================================================
-- Favorites & recents
-- =============================================================================

CREATE TABLE study_guide_favorites (
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    study_guide_id UUID        NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, study_guide_id)
);

CREATE INDEX idx_study_guide_favorites_user_created
    ON study_guide_favorites(user_id, created_at, study_guide_id);

CREATE TABLE study_guide_last_viewed (
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    study_guide_id UUID        NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
    viewed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, study_guide_id)
);

CREATE INDEX idx_study_guide_last_viewed_user_viewed
    ON study_guide_last_viewed(user_id, viewed_at, study_guide_id);

CREATE TABLE course_favorites (
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id  UUID        NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, course_id)
);

CREATE INDEX idx_course_favorites_user_created
    ON course_favorites(user_id, created_at, course_id);

CREATE TABLE course_last_viewed (
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID        NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, course_id)
);

CREATE INDEX idx_course_last_viewed_user_viewed
    ON course_last_viewed(user_id, viewed_at, course_id);
