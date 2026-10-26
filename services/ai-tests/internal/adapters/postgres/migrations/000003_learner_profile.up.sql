ALTER TABLE adaptive_questions ADD COLUMN skill_code TEXT NOT NULL DEFAULT '';
ALTER TABLE adaptive_rounds ADD COLUMN learning_plan JSONB NOT NULL DEFAULT '[]';
CREATE TABLE learner_skill_states (
 user_id BIGINT NOT NULL,
 skill_code TEXT NOT NULL,
 state JSONB NOT NULL,
 PRIMARY KEY(user_id,skill_code)
);
CREATE TABLE learning_evidence (
 question_id BIGINT PRIMARY KEY REFERENCES adaptive_questions(id) ON DELETE CASCADE,
 session_id UUID NOT NULL REFERENCES adaptive_sessions(id) ON DELETE CASCADE,
 user_id BIGINT NOT NULL,
 skill_code TEXT NOT NULL,
 fingerprint CHAR(64) NOT NULL,
 is_correct BOOLEAN NOT NULL,
 difficulty SMALLINT NOT NULL CHECK(difficulty BETWEEN 1 AND 5),
 contributes BOOLEAN NOT NULL,
 policy_version TEXT NOT NULL,
 recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX learning_evidence_novel ON learning_evidence(user_id,skill_code,fingerprint) WHERE contributes;
CREATE INDEX learning_evidence_sessions ON learning_evidence(user_id,skill_code,session_id) WHERE contributes;
