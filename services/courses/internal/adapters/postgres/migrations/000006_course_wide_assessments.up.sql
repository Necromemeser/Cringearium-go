-- Make pre/post assessments cover linear, quadratic, and trigonometric equations.
-- Keep the existing eight trigonometry questions and add eight questions for each
-- of the other two course modules, resulting in a balanced 24-question assessment.

-- Preserve the scope of historical attempts made with the original trig-only tests.
ALTER TABLE assessment_results
    ADD COLUMN assessment_scope TEXT NOT NULL DEFAULT 'course'
    CHECK (assessment_scope IN ('trigonometry', 'course'));

UPDATE assessment_results ar
SET assessment_scope = 'trigonometry'
FROM course_pages p
WHERE ar.page_id = p.id
  AND p.title IN ('Входной тест: тригонометрия', 'Итоговый тест: тригонометрия');

-- Reserve position 0 for the entry assessment while preserving section order.
-- The temporary offset avoids collisions if (course_id, position) is unique.
UPDATE course_sections
SET position = position + 100000
WHERE course_id IN (
    SELECT id FROM courses
    WHERE title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
);

UPDATE course_sections
SET position = position - 99999
WHERE course_id IN (
    SELECT id FROM courses
    WHERE title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
);

INSERT INTO course_sections (course_id, title, description, position)
SELECT c.id, 'Входное тестирование',
       'Проверка исходных знаний по линейным, квадратным и тригонометрическим уравнениям.',
       0
FROM courses c
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO course_sections (course_id, title, description, position)
SELECT c.id, 'Итоговое тестирование',
       'Итоговая проверка знаний по линейным, квадратным и тригонометрическим уравнениям.',
       COALESCE((SELECT MAX(s.position) + 1 FROM course_sections s WHERE s.course_id = c.id), 1)
FROM courses c
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

-- Keep the same page/test IDs so saved attempts and assessment-result references survive.
UPDATE course_pages p
SET section_id = target_section.id,
    title = 'Входной тест: весь курс',
    position = 0
FROM course_sections old_section
JOIN courses c ON c.id = old_section.course_id
JOIN course_sections target_section
  ON target_section.course_id = c.id
 AND target_section.title = 'Входное тестирование'
WHERE p.section_id = old_section.id
  AND old_section.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
  AND p.title = 'Входной тест: тригонометрия';

UPDATE course_pages p
SET section_id = target_section.id,
    title = 'Итоговый тест: весь курс',
    position = 0
FROM course_sections old_section
JOIN courses c ON c.id = old_section.course_id
JOIN course_sections target_section
  ON target_section.course_id = c.id
 AND target_section.title = 'Итоговое тестирование'
WHERE p.section_id = old_section.id
  AND old_section.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
  AND p.title = 'Итоговый тест: тригонометрия';

CREATE TEMP TABLE course_assessment_questions (
    assessment_title TEXT NOT NULL,
    question TEXT NOT NULL,
    question_position INTEGER NOT NULL,
    options TEXT[] NOT NULL,
    correct_position INTEGER NOT NULL
) ON COMMIT DROP;

-- Entry assessment: eight linear-equation and eight quadratic-equation questions.
INSERT INTO course_assessment_questions VALUES
('Входной тест: весь курс','Решите уравнение: 3x − 7 = 11.',8,ARRAY['x = 4','x = −6','x = 18','x = 6'],4),
('Входной тест: весь курс','Решите уравнение: 2(x + 3) = 14.',9,ARRAY['x = 7','x = 5','x = 4','x = 10'],3),
('Входной тест: весь курс','Решите уравнение: 5x + 4 = 2x + 19.',10,ARRAY['x = 3','x = 5','x = 15','x = −5'],2),
('Входной тест: весь курс','Решите уравнение: (x − 2) / 3 = 4.',11,ARRAY['x = 14','x = 10','x = 6','x = 12'],1),
('Входной тест: весь курс','Сколько решений имеет уравнение 0x = 5?',12,ARRAY['Бесконечно много','Только x = 0','Только x = 5','Ни одного'],4),
('Входной тест: весь курс','Решите уравнение: 4x − 3 = 2x + 9.',13,ARRAY['x = 3','x = 9','x = 6','x = −6'],3),
('Входной тест: весь курс','Решите уравнение: 3(x − 2) = 2x + 1.',14,ARRAY['x = 5','x = 7','x = 3','x = −7'],2),
('Входной тест: весь курс','Сколько решений имеет уравнение 2x + 5 = 2x + 5?',15,ARRAY['Бесконечно много','Ни одного','Только x = 0','Только x = 5'],1),
('Входной тест: весь курс','Найдите корни уравнения x² − 9 = 0.',16,ARRAY['x = 0 и x = 9','Только x = 3','Действительных корней нет','x = −3 и x = 3'],4),
('Входной тест: весь курс','Чему равен дискриминант уравнения x² − 5x + 6 = 0?',17,ARRAY['25','−1','1','0'],3),
('Входной тест: весь курс','Что можно сказать о действительных корнях, если D < 0?',18,ARRAY['Всегда два корня','Действительных корней нет','Всегда один корень','Оба корня равны нулю'],2),
('Входной тест: весь курс','Найдите корни уравнения x² − 5x + 6 = 0.',19,ARRAY['x = 2 и x = 3','x = −2 и x = −3','x = 1 и x = 6','Действительных корней нет'],1),
('Входной тест: весь курс','Найдите корни уравнения 2x² − 8 = 0.',20,ARRAY['x = 0 и x = 4','x = −4 и x = 4','Только x = 2','x = −2 и x = 2'],4),
('Входной тест: весь курс','Чему равна сумма корней уравнения x² − 5x + 6 = 0?',21,ARRAY['6','−5','5','1'],3),
('Входной тест: весь курс','Решите уравнение x² + 6x + 9 = 0.',22,ARRAY['x = 3 и x = −3','x = −3','x = 9','Действительных корней нет'],2),
('Входной тест: весь курс','Найдите корни уравнения x² − 7x + 12 = 0.',23,ARRAY['x = 3 и x = 4','x = −3 и x = −4','x = 2 и x = 6','Действительных корней нет'],1),

-- Final assessment: equivalent skills with different values and formulations.
('Итоговый тест: весь курс','Решите уравнение: 4x + 1 = 13.',8,ARRAY['x = 4','x = 12','x = −3','x = 3'],4),
('Итоговый тест: весь курс','Решите уравнение: 3(x + 2) = 21.',9,ARRAY['x = 7','x = 3','x = 5','x = 9'],3),
('Итоговый тест: весь курс','Решите уравнение: 7x − 2 = 4x + 13.',10,ARRAY['x = 3','x = 5','x = 15','x = −5'],2),
('Итоговый тест: весь курс','Решите уравнение: (x + 5) / 2 = 6.',11,ARRAY['x = 7','x = 12','x = 17','x = 1'],1),
('Итоговый тест: весь курс','Сколько решений имеет уравнение 0x = 0?',12,ARRAY['Ни одного','Только x = 0','Только x = 1','Бесконечно много'],4),
('Итоговый тест: весь курс','Решите уравнение: 6x + 5 = 2x + 21.',13,ARRAY['x = 6','x = 16','x = 4','x = −4'],3),
('Итоговый тест: весь курс','Решите уравнение: 2(x − 4) = x + 3.',14,ARRAY['x = 7','x = 11','x = −11','x = 5'],2),
('Итоговый тест: весь курс','Сколько решений имеет уравнение 3x + 2 = 3x − 4?',15,ARRAY['Ни одного','Бесконечно много','Только x = 0','Только x = −2'],1),
('Итоговый тест: весь курс','Найдите корни уравнения x² − 16 = 0.',16,ARRAY['x = −8 и x = 8','Только x = 4','Действительных корней нет','x = −4 и x = 4'],4),
('Итоговый тест: весь курс','Чему равен дискриминант уравнения x² + 2x + 1 = 0?',17,ARRAY['4','−4','0','1'],3),
('Итоговый тест: весь курс','Каковы действительные корни уравнения x² + 2x + 5 = 0?',18,ARRAY['Два различных корня','Действительных корней нет','Один корень x = −1','Два равных корня x = 1'],2),
('Итоговый тест: весь курс','Найдите корни уравнения x² − 9x + 20 = 0.',19,ARRAY['x = 4 и x = 5','x = −4 и x = −5','x = 2 и x = 10','Действительных корней нет'],1),
('Итоговый тест: весь курс','Найдите корни уравнения 3x² − 12 = 0.',20,ARRAY['x = −4 и x = 4','x = 0 и x = 4','Только x = 2','x = −2 и x = 2'],4),
('Итоговый тест: весь курс','Чему равна сумма корней уравнения x² + 7x + 10 = 0?',21,ARRAY['7','10','−7','−10'],3),
('Итоговый тест: весь курс','Решите уравнение x² − 6x + 9 = 0.',22,ARRAY['x = −3','x = 3','x = 3 и x = −3','Действительных корней нет'],2),
('Итоговый тест: весь курс','Найдите корни уравнения x² − 2x − 8 = 0.',23,ARRAY['x = −2 и x = 4','x = 2 и x = −4','x = −4 и x = 2','Действительных корней нет'],1);

INSERT INTO test_questions (test_id, question, position)
SELECT t.id, b.question, b.question_position
FROM course_assessment_questions b
JOIN course_pages p ON p.title = b.assessment_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO test_answers (question_id, text, is_correct, position)
SELECT q.id, option_data.option_text,
       option_data.option_position = b.correct_position,
       option_data.option_position - 1
FROM course_assessment_questions b
JOIN course_pages p ON p.title = b.assessment_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
JOIN test_questions q ON q.test_id = t.id AND q.question = b.question
CROSS JOIN LATERAL unnest(b.options) WITH ORDINALITY AS option_data(option_text, option_position)
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');
