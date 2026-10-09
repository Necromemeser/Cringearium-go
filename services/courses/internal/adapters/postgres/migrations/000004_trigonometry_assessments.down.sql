DELETE FROM test_answers
WHERE question_id IN (
    SELECT q.id
    FROM test_questions q
    JOIN tests t ON t.id = q.test_id
    JOIN course_pages p ON p.id = t.page_id
    WHERE p.title IN ('Входной тест: тригонометрия', 'Итоговый тест: тригонометрия')
);

DELETE FROM test_questions
WHERE test_id IN (
    SELECT t.id
    FROM tests t
    JOIN course_pages p ON p.id = t.page_id
    WHERE p.title IN ('Входной тест: тригонометрия', 'Итоговый тест: тригонометрия')
);

DELETE FROM tests
WHERE page_id IN (
    SELECT p.id
    FROM course_pages p
    WHERE p.title IN ('Входной тест: тригонометрия', 'Итоговый тест: тригонометрия')
);

DELETE FROM course_pages
WHERE title IN (
    'Входной тест: тригонометрия',
    'Итоговый тест: тригонометрия',
    'Таблица значений и знаки функций',
    'Основные тождества и преобразования',
    'Алгоритм решения: от уравнения к ответу'
);

UPDATE course_pages p
SET content = CASE p.title
    WHEN 'Тригонометрия и единичная окружность' THEN
'# Тригонометрические уравнения

Решением тригонометрического уравнения может быть целое семейство значений. В формулах n обозначает любое целое число, а углы измеряются в радианах.

- sin x и cos x имеют период 2π;
- tg x имеет период π.

Единичная окружность помогает понять, где синус и косинус принимают заданные значения.'
    WHEN 'Базовые тригонометрические уравнения' THEN
'# Базовые решения

- sin x = 0 → x = πn;
- cos x = 1 → x = 2πn;
- tg x = 0 → x = πn,

где n — любое целое число.

У синуса и косинуса решения повторяются через 2π, у тангенса — через π. При решении на заданном промежутке из общего семейства нужно выбрать подходящие значения.'
    WHEN 'Решения на промежутке' THEN
'# Решения на заданном промежутке

Сначала найдите общее решение, затем подберите целые значения n, которые попадают в нужный промежуток.

**Пример:** cos x = 1 на отрезке [0; 2π]. Общее решение: x = 2πn. На указанном отрезке подходят x = 0 и x = 2π.

Проверяйте, включены ли границы промежутка: это зависит от записи скобок.'
    ELSE p.content
END
FROM course_sections s
JOIN courses c ON c.id = s.course_id
WHERE p.section_id = s.id
  AND s.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

UPDATE course_pages p
SET position = CASE p.title
    WHEN 'Тригонометрия и единичная окружность' THEN 0
    WHEN 'Базовые тригонометрические уравнения' THEN 1
    WHEN 'Решения на промежутке' THEN 2
    WHEN 'Проверка: тригонометрические уравнения' THEN 3
    WHEN 'Практика с обратной связью I: тригонометрия' THEN 4
    WHEN 'Практика с обратной связью II: тригонометрия' THEN 5
    ELSE p.position
END
FROM course_sections s
JOIN courses c ON c.id = s.course_id
WHERE p.section_id = s.id
  AND s.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');
