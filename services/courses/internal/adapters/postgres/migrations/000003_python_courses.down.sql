-- Remove seeded Python courses; dependent content cascades automatically.
DELETE FROM courses
WHERE title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

ALTER TABLE tests ALTER COLUMN passing_score SET DEFAULT 50;
