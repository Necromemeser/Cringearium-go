-- Removing the math seed removes all dependent sections, pages, tests and answers.
DROP TABLE IF EXISTS assessment_results;
DELETE FROM courses
WHERE title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');
