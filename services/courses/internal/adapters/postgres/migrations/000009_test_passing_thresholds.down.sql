-- Restore the previous global passing threshold.
ALTER TABLE tests
    ALTER COLUMN passing_score SET DEFAULT 70;

UPDATE tests
SET passing_score = 70;
