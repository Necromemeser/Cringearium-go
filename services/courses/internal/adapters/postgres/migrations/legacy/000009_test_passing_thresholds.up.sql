-- Entry and final tests are diagnostic/completion checks, not pass/fail gates.
-- All other regular tests require at least 50%.
ALTER TABLE tests
    ALTER COLUMN passing_score SET DEFAULT 50;

UPDATE tests t
SET passing_score = CASE
    WHEN p.title ILIKE 'Входн%' OR p.title ILIKE 'Итогов%' THEN 0
    ELSE 50
END
FROM course_pages p
WHERE p.id = t.page_id;
