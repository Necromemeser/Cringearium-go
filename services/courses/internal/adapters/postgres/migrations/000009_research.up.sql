CREATE TABLE assessment_definitions (
 test_id BIGINT PRIMARY KEY REFERENCES tests(id) ON DELETE CASCADE,
 kind TEXT NOT NULL CHECK(kind IN ('pretest','posttest')),
 scope TEXT NOT NULL CHECK(scope IN ('course','trigonometry'))
);
INSERT INTO assessment_definitions(test_id,kind,scope)
SELECT t.id,CASE WHEN p.title LIKE 'Входной тест:%' THEN 'pretest' ELSE 'posttest' END,
 CASE WHEN p.title LIKE '%тригонометрия%' THEN 'trigonometry' ELSE 'course' END
FROM tests t JOIN course_pages p ON p.id=t.page_id
WHERE p.title LIKE 'Входной тест:%' OR p.title LIKE 'Итоговый тест:%';
ALTER TABLE assessment_results ADD COLUMN form_hash TEXT;
ALTER TABLE assessment_results ADD COLUMN instrument_snapshot JSONB;
-- Historical versions cannot be reconstructed reliably; leave them explicitly unknown.
CREATE TABLE research_studies (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 title TEXT NOT NULL,
 design TEXT NOT NULL CHECK(design IN ('observational','randomized')),
 scope TEXT NOT NULL CHECK(scope IN ('course','trigonometry')),
 control_course_id BIGINT NOT NULL REFERENCES courses(id),
 adaptive_course_id BIGINT NOT NULL REFERENCES courses(id),
 pre_hash TEXT NOT NULL,
 post_hash TEXT NOT NULL,
 content_hash TEXT NOT NULL,
 protocol_notes TEXT NOT NULL DEFAULT '',
 export_salt TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK(control_course_id<>adaptive_course_id)
);
CREATE TABLE research_participants (
 study_id BIGINT NOT NULL REFERENCES research_studies(id),
 user_id BIGINT NOT NULL,
 arm TEXT NOT NULL CHECK(arm IN ('control','adaptive')),
 assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY(study_id,user_id)
);
CREATE INDEX research_participant_user ON research_participants(user_id,study_id);
CREATE INDEX assessment_research ON assessment_results(user_id,course_id,assessment_scope,completed_at,id);
