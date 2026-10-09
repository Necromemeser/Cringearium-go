-- Remove the course-wide questions and restore the original trigonometry-only placement.

DELETE FROM test_answers a
USING test_questions q, tests t, course_pages p, course_sections s, courses c
WHERE a.question_id = q.id
  AND q.test_id = t.id
  AND t.page_id = p.id
  AND p.section_id = s.id
  AND s.course_id = c.id
  AND p.title IN ('Входной тест: весь курс', 'Итоговый тест: весь курс')
  AND q.position >= 8
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

DELETE FROM test_questions q
USING tests t, course_pages p, course_sections s, courses c
WHERE q.test_id = t.id
  AND t.page_id = p.id
  AND p.section_id = s.id
  AND s.course_id = c.id
  AND p.title IN ('Входной тест: весь курс', 'Итоговый тест: весь курс')
  AND q.position >= 8
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

UPDATE course_pages p
SET section_id = original_section.id,
    title = CASE
        WHEN p.title = 'Входной тест: весь курс' THEN 'Входной тест: тригонометрия'
        ELSE 'Итоговый тест: тригонометрия'
    END,
    position = CASE
        WHEN p.title = 'Входной тест: весь курс' THEN 6
        ELSE 10
    END
FROM course_sections assessment_section
JOIN courses c ON c.id = assessment_section.course_id
JOIN course_sections original_section
  ON original_section.course_id = c.id
 AND original_section.title = 'Тригонометрические уравнения'
WHERE p.section_id = assessment_section.id
  AND assessment_section.title IN ('Входное тестирование', 'Итоговое тестирование')
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

DELETE FROM course_sections
WHERE title IN ('Входное тестирование', 'Итоговое тестирование')
  AND course_id IN (
      SELECT id FROM courses
      WHERE title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
  );

UPDATE course_sections
SET position = position - 1
WHERE course_id IN (
    SELECT id FROM courses
    WHERE title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
  )
  AND title NOT IN ('Входное тестирование', 'Итоговое тестирование');

ALTER TABLE assessment_results DROP COLUMN assessment_scope;
