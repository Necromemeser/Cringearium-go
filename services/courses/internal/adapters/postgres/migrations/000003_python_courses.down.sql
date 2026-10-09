-- Remove both Python courses and restore the math thresholds.
DELETE FROM courses
WHERE title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

UPDATE tests t
SET passing_score = CASE
    WHEN p.title ILIKE 'Входн%' OR p.title ILIKE 'Итогов%' THEN 0
    ELSE 50
END
FROM course_pages p
WHERE p.id = t.page_id;

ALTER TABLE tests ALTER COLUMN passing_score SET DEFAULT 50;
