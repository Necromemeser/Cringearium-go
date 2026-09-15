CREATE TABLE tests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    course_id BIGINT NOT NULL,
    material_id BIGINT,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE attempts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    score INTEGER,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_tests_course_id ON tests(course_id);
CREATE INDEX idx_questions_test_id ON questions(test_id);
CREATE INDEX idx_attempts_test_id ON attempts(test_id);
CREATE INDEX idx_attempts_user_id ON attempts(user_id);