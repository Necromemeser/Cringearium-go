CREATE TABLE assessment_results (
    id BIGSERIAL PRIMARY KEY,
    test_attempt_id BIGINT NOT NULL UNIQUE REFERENCES test_attempts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    page_id BIGINT NOT NULL REFERENCES course_pages(id) ON DELETE CASCADE,
    assessment_type TEXT NOT NULL CHECK (assessment_type IN ('pretest', 'posttest')),
    score INTEGER NOT NULL CHECK (score BETWEEN 0 AND 100),
    answers JSONB NOT NULL DEFAULT '[]'::jsonb,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX assessment_results_user_course_idx
    ON assessment_results (user_id, course_id, assessment_type);

CREATE INDEX assessment_results_completed_at_idx
    ON assessment_results (completed_at DESC);
