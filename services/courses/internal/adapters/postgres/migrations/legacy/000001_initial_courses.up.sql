CREATE TYPE course_status AS ENUM (
    'draft',
    'published',
    'archived'
);

CREATE TYPE page_type AS ENUM (
    'theory',
    'test',
    'ai_test'
);

CREATE TABLE courses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    theme VARCHAR(255),
    description TEXT,
    price INTEGER NOT NULL DEFAULT 0,
    image_id VARCHAR(255),
    author_id BIGINT,
    status course_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (price >= 0)
);

CREATE TABLE course_sections (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (course_id, position),
    CHECK (position >= 0)
);

CREATE TABLE course_pages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id BIGINT NOT NULL REFERENCES course_sections(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type page_type NOT NULL,
    content TEXT,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (section_id, position),
    CHECK (position >= 0)
);

CREATE TABLE tests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    page_id BIGINT NOT NULL UNIQUE REFERENCES course_pages(id) ON DELETE CASCADE,
    passing_score INTEGER NOT NULL DEFAULT 70,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (passing_score BETWEEN 0 AND 100)
);

CREATE TABLE test_questions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    position INTEGER NOT NULL,

    UNIQUE (test_id, position),
    CHECK (position >= 0)
);

CREATE TABLE test_answers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES test_questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    position INTEGER NOT NULL,

    UNIQUE (question_id, position),
    CHECK (position >= 0)
);

CREATE TABLE course_access (
    user_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, course_id)
);

CREATE TABLE page_progress (
    user_id BIGINT NOT NULL,
    page_id BIGINT NOT NULL REFERENCES course_pages(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, page_id)
);

CREATE TABLE test_attempts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    score INTEGER,
    passed BOOLEAN NOT NULL DEFAULT FALSE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    CHECK (score IS NULL OR score BETWEEN 0 AND 100),
    CHECK (
        completed_at IS NULL
        OR completed_at >= started_at
    )
);

CREATE TABLE test_attempt_answers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES test_questions(id) ON DELETE CASCADE,
    answer_id BIGINT REFERENCES test_answers(id) ON DELETE SET NULL,
    is_correct BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (attempt_id, question_id)
);

CREATE INDEX idx_course_sections_course_id
    ON course_sections(course_id);

CREATE INDEX idx_course_pages_section_id
    ON course_pages(section_id);

CREATE INDEX idx_tests_page_id
    ON tests(page_id);

CREATE INDEX idx_test_questions_test_id
    ON test_questions(test_id);

CREATE INDEX idx_test_answers_question_id
    ON test_answers(question_id);

CREATE INDEX idx_course_access_course_id
    ON course_access(course_id);

CREATE INDEX idx_page_progress_page_id
    ON page_progress(page_id);

CREATE INDEX idx_test_attempts_test_id
    ON test_attempts(test_id);

CREATE INDEX idx_test_attempts_user_id
    ON test_attempts(user_id);

CREATE INDEX idx_test_attempt_answers_attempt_id
    ON test_attempt_answers(attempt_id);

CREATE INDEX idx_test_attempt_answers_question_id
    ON test_attempt_answers(question_id);