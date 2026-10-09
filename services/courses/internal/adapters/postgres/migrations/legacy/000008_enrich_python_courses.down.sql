-- Remove the additional questions and restore the original course descriptions/content.

DELETE FROM test_answers a
USING test_questions q, tests t, course_pages p, course_sections s, courses c
WHERE a.question_id = q.id
  AND q.test_id = t.id
  AND t.page_id = p.id
  AND p.section_id = s.id
  AND s.course_id = c.id
  AND c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования')
  AND q.question IN (
      'Что вернёт len("кот")?',
      'Какой символ начинает однострочный комментарий в Python?',
      'Что вернёт int("12")?',
      'Какой оператор возводит число в степень?',
      'Что делает функция print()?',
      'Что вернёт len([4, 7, 9])?',
      'Что вернёт выражение [10, 20, 30][1]?',
      'Чему равно выражение 3 ** 2?',
      'Что выведет код: for i in range(2, 5): print(i)?',
      'Что вернёт вызов double(4), если функция объявлена как def double(x): return x * 2?',
      'Что означает выражение not (4 > 2)?',
      'Что вернёт sum([1, 2, 3])?'
  );

DELETE FROM test_questions q
USING tests t, course_pages p, course_sections s, courses c
WHERE q.test_id = t.id
  AND t.page_id = p.id
  AND p.section_id = s.id
  AND s.course_id = c.id
  AND c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования')
  AND q.question IN (
      'Что вернёт len("кот")?',
      'Какой символ начинает однострочный комментарий в Python?',
      'Что вернёт int("12")?',
      'Какой оператор возводит число в степень?',
      'Что делает функция print()?',
      'Что вернёт len([4, 7, 9])?',
      'Что вернёт выражение [10, 20, 30][1]?',
      'Чему равно выражение 3 ** 2?',
      'Что выведет код: for i in range(2, 5): print(i)?',
      'Что вернёт вызов double(4), если функция объявлена как def double(x): return x * 2?',
      'Что означает выражение not (4 > 2)?',
      'Что вернёт sum([1, 2, 3])?'
  );

UPDATE courses
SET description = CASE title
    WHEN 'Python: основы программирования' THEN
        'Базовый курс для первого знакомства с Python: переменные, условия, циклы, коллекции и функции. Теория, обычные тесты и итоговая проверка.'
    WHEN 'Python с ИИ: основы программирования' THEN
        'Те же темы и контрольные тесты, дополненные адаптивной практикой с ИИ-тестами по текущему материалу.'
    ELSE description
END
WHERE title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

UPDATE course_sections s
SET description = v.description
FROM (VALUES
    ('Основы Python', 'Разберём переменные, типы данных, ввод и вывод.'),
    ('Условия и циклы', 'Научимся выбирать действия и повторять их с помощью циклов.'),
    ('Коллекции и функции', 'Сгруппируем данные и научимся выделять повторяемый код в функции.')
) AS v(title, description)
JOIN courses c ON c.id = s.course_id
WHERE s.title = v.title
  AND c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

UPDATE course_pages p
SET content = left(p.content, strpos(p.content, v.marker) - 1)
FROM (VALUES
    ('Переменные и типы данных', E'\n\n## Как думать о типах данных'),
    ('Ввод, вывод и операции', E'\n\n## Как составить небольшую программу'),
    ('Условия if, elif и else', E'\n\n## Сначала простое условие, затем более конкретное'),
    ('Циклы for и while', E'\n\n## Выбираем подходящий цикл'),
    ('Списки и словари', E'\n\n## Изменение и перебор коллекций'),
    ('Функции и возвращаемые значения', E'\n\n## Аргументы и результат')
) AS v(title, marker)
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE p.title = v.title
  AND p.type = 'theory'
  AND c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования')
  AND strpos(p.content, v.marker) > 0;
