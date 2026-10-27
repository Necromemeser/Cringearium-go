DROP INDEX assessment_research;
DROP TABLE research_participants;
DROP TABLE research_studies;
ALTER TABLE assessment_results DROP COLUMN instrument_snapshot;
ALTER TABLE assessment_results DROP COLUMN form_hash;
DROP TABLE assessment_definitions;
