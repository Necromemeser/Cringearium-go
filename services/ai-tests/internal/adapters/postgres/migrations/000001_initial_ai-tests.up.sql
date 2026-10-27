CREATE TABLE adaptive_sessions (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    topic_page_id BIGINT,

    status VARCHAR(20) NOT NULL DEFAULT 'in_progress'
        CHECK (status IN (
            'in_progress',
            'completed',
            'failed'
        )),

    current_round INTEGER NOT NULL DEFAULT 1
        CHECK (current_round BETWEEN 1 AND 2),

    question_count INTEGER NOT NULL
        CHECK (question_count BETWEEN 3 AND 10),

    context_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE adaptive_rounds (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id UUID NOT NULL
        REFERENCES adaptive_sessions(id) ON DELETE CASCADE,

    round_number INTEGER NOT NULL
        CHECK (round_number BETWEEN 1 AND 2),

    strategy VARCHAR(30) NOT NULL
        CHECK (strategy IN (
            'initial',
            'increase_difficulty',
            'targeted_practice'
        )),

    status VARCHAR(20) NOT NULL DEFAULT 'generated'
        CHECK (status IN (
            'generating',
            'generated',
            'completed',
            'failed'
        )),

    title VARCHAR(255) NOT NULL,
    instructions TEXT,

    model VARCHAR(100),
    prompt_version VARCHAR(50),
    prompt_hash VARCHAR(64),

    input_tokens INTEGER,
    output_tokens INTEGER,

    generation_error TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    UNIQUE (session_id, round_number),
    UNIQUE (id, session_id)
);

CREATE TABLE adaptive_questions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    round_id BIGINT NOT NULL
        REFERENCES adaptive_rounds(id) ON DELETE CASCADE,

    position INTEGER NOT NULL CHECK (position > 0),

    topic_page_id BIGINT,
    topic_title VARCHAR(255),

    question TEXT NOT NULL,
    difficulty SMALLINT NOT NULL
        CHECK (difficulty BETWEEN 1 AND 5),

    correct_option_key VARCHAR(20) NOT NULL,
    explanation TEXT NOT NULL,

    knowledge_basis VARCHAR(30) NOT NULL
        CHECK (knowledge_basis IN (
            'course',
            'external_knowledge',
            'mixed'
        )),

    sources JSONB NOT NULL DEFAULT '[]'::JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (round_id, position),
    UNIQUE (id, round_id)
);

CREATE TABLE adaptive_options (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id BIGINT NOT NULL
        REFERENCES adaptive_questions(id) ON DELETE CASCADE,

    option_key VARCHAR(20) NOT NULL,
    text TEXT NOT NULL,
    position INTEGER NOT NULL CHECK (position > 0),

    UNIQUE (question_id, option_key),
    UNIQUE (question_id, position)
);

CREATE TABLE adaptive_answers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    session_id UUID NOT NULL
        REFERENCES adaptive_sessions(id) ON DELETE CASCADE,

    round_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,

    selected_option_key VARCHAR(20) NOT NULL,
    is_correct BOOLEAN NOT NULL,

    answered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (round_id, session_id)
        REFERENCES adaptive_rounds(id, session_id)
        ON DELETE CASCADE,

    FOREIGN KEY (question_id, round_id)
        REFERENCES adaptive_questions(id, round_id)
        ON DELETE CASCADE,

    FOREIGN KEY (question_id, selected_option_key)
        REFERENCES adaptive_options(question_id, option_key),

    UNIQUE (question_id)
);

CREATE TABLE adaptive_feedback (
    session_id UUID PRIMARY KEY
        REFERENCES adaptive_sessions(id) ON DELETE CASCADE,

    summary TEXT NOT NULL,

    mastered_topics JSONB NOT NULL DEFAULT '[]'::JSONB,
    topics_to_review JSONB NOT NULL DEFAULT '[]'::JSONB,
    next_steps JSONB NOT NULL DEFAULT '[]'::JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_adaptive_sessions_user
    ON adaptive_sessions(user_id, created_at DESC);

CREATE INDEX idx_adaptive_sessions_course
    ON adaptive_sessions(course_id);

CREATE INDEX idx_adaptive_rounds_session
    ON adaptive_rounds(session_id);

CREATE INDEX idx_adaptive_questions_round
    ON adaptive_questions(round_id);

CREATE INDEX idx_adaptive_answers_session
    ON adaptive_answers(session_id);

CREATE INDEX idx_adaptive_answers_round
    ON adaptive_answers(round_id);