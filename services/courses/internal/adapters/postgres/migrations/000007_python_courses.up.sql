-- Seed two introductory Python courses with the same three-section syllabus.
-- The AI variant adds adaptive practice pages; both courses keep fixed entry, section and final tests.

INSERT INTO courses (title, theme, description, price, image_id, author_id, status)
VALUES
('Python: основы программирования', 'Python', 'Базовый курс для первого знакомства с Python: переменные, условия, циклы, коллекции и функции. Теория, обычные тесты и итоговая проверка.', 0, '/images/python-basics.svg', NULL, 'published'),
('Python с ИИ: основы программирования', 'Python + ИИ', 'Те же темы и контрольные тесты, дополненные адаптивной практикой с ИИ-тестами по текущему материалу.', 0, '/images/python-ai.svg', NULL, 'published');

INSERT INTO course_sections (course_id, title, description, position)
SELECT c.id, s.title, s.description, s.position
FROM courses c
CROSS JOIN (VALUES
('Основы Python', 'Разберём переменные, типы данных, ввод и вывод.', 0),
('Условия и циклы', 'Научимся выбирать действия и повторять их с помощью циклов.', 1),
('Коллекции и функции', 'Сгруппируем данные и научимся выделять повторяемый код в функции.', 2)
) AS s(title, description, position)
WHERE c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

INSERT INTO course_pages (section_id, title, type, content, position)
SELECT s.id, p.title, p.page_type::page_type, p.content, p.position
FROM courses c
JOIN course_sections s ON s.course_id = c.id
CROSS JOIN (VALUES
('Основы Python', 'Входной тест: основы Python', 'test', NULL, 0, false),
('Основы Python', 'Переменные и типы данных', 'theory', '# Переменные и типы данных

Переменная — имя, по которому программа обращается к сохранённому значению. В Python значение присваивают с помощью знака `=`.

```python
name = "Мира"
age = 16
height = 1.68
is_student = True
```

Здесь `name` хранит строку (`str`), `age` — целое число (`int`), `height` — дробное число (`float`), а `is_student` — логическое значение (`bool`).

![Примеры переменных и типов данных в Python](/images/python-variables.svg)

Имена переменных могут содержать буквы, цифры и подчёркивания, но не должны начинаться с цифры. Регистр важен: `score` и `Score` — разные имена.

Чтобы узнать тип значения, используйте `type(value)`.', 1, false),
('Основы Python', 'Ввод, вывод и операции', 'theory', '# Ввод, вывод и арифметика

Функция `print()` выводит данные, а `input()` читает строку, введённую пользователем.

```python
name = input("Как тебя зовут? ")
print("Привет,", name)
```

Даже если пользователь вводит цифры, `input()` возвращает строку. Для вычислений преобразуйте её, например: `age = int(input("Возраст: "))`.

Основные арифметические операции:
- `+` сложение;
- `-` вычитание;
- `*` умножение;
- `/` деление;
- `//` целочисленное деление;
- `%` остаток от деления;
- `**` возведение в степень.

**Пример:** `17 // 5` даст `3`, а `17 % 5` даст `2`.', 2, false),
('Основы Python', 'Проверка: основы Python', 'test', NULL, 3, false),
('Основы Python', 'Практика с ИИ: переменные и ввод', 'ai_test', NULL, 4, true),
('Условия и циклы', 'Условия if, elif и else', 'theory', '# Условия

Условная конструкция позволяет выполнить разные действия в зависимости от истинности выражения.

```python
temperature = 12

if temperature >= 20:
    print("Тепло")
elif temperature >= 10:
    print("Прохладно")
else:
    print("Холодно")
```

Python использует отступы, чтобы обозначить блок кода. Все команды внутри одной ветки должны иметь одинаковый отступ.

Операторы сравнения: `==` равно, `!=` не равно, `>`, `<`, `>=`, `<=`. Не путайте присваивание `=` со сравнением `==`.', 0, false),
('Условия и циклы', 'Циклы for и while', 'theory', '# Повторение действий

Цикл `for` перебирает элементы последовательности. Функция `range(stop)` создаёт числа от нуля до `stop - 1`.

```python
for number in range(1, 4):
    print(number)
```

Этот код выведет 1, 2 и 3.

Цикл `while` повторяется, пока условие истинно:

```python
count = 3
while count > 0:
    print(count)
    count -= 1
```

Следите, чтобы условие `while` когда-нибудь стало ложным, иначе цикл может стать бесконечным. `break` завершает цикл, а `continue` переходит к следующей итерации.

![Схема выбора условия и повторения в программе](/images/python-flow.svg)', 1, false),
('Условия и циклы', 'Проверка: условия и циклы', 'test', NULL, 2, false),
('Условия и циклы', 'Практика с ИИ: условия и циклы', 'ai_test', NULL, 3, true),
('Коллекции и функции', 'Списки и словари', 'theory', '# Коллекции

Список хранит упорядоченный набор значений. Индексы начинаются с нуля.

```python
scores = [8, 10, 7]
scores.append(9)
print(scores[0])  # 8
```

Словарь хранит пары «ключ — значение»:

```python
student = {"name": "Мира", "grade": 10}
print(student["name"])
```

Для перебора элементов используют `for`. Метод `append()` добавляет элемент в список, а обращение к словарю по ключу возвращает соответствующее значение.', 0, false),
('Коллекции и функции', 'Функции и возвращаемые значения', 'theory', '# Функции

Функция объединяет инструкции, которые можно вызывать повторно. Её объявляют с помощью `def`.

```python
def square(number):
    return number * number

result = square(5)
print(result)  # 25
```

Параметр `number` получает значение при вызове. `return` возвращает результат в место вызова; `print()` только выводит значение на экран.

Функции помогают разбить программу на небольшие части, давать действиям понятные имена и избегать повторения одного и того же кода.', 1, false),
('Коллекции и функции', 'Проверка: коллекции и функции', 'test', NULL, 2, false),
('Коллекции и функции', 'Практика с ИИ: коллекции и функции', 'ai_test', NULL, 3, true),
('Коллекции и функции', 'Итоговый тест: основы Python', 'test', NULL, 4, false)
) AS p(section_title, title, page_type, content, position, ai_only)
WHERE c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования')
  AND s.title = p.section_title
  AND (NOT p.ai_only OR c.title = 'Python с ИИ: основы программирования');

INSERT INTO tests (page_id, passing_score)
SELECT p.id, 70
FROM course_pages p
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE p.type = 'test'
  AND c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

CREATE TEMP TABLE python_question_bank (
    page_title TEXT NOT NULL,
    question TEXT NOT NULL,
    question_position INTEGER NOT NULL,
    options TEXT[] NOT NULL,
    correct_position INTEGER NOT NULL
) ON COMMIT DROP;

INSERT INTO python_question_bank VALUES
('Входной тест: основы Python', 'Какой знак используется для присваивания значения переменной?', 0, ARRAY['=', '==', ':=', '=>'], 1),
('Входной тест: основы Python', 'Какой тип данных хранит текст?', 1, ARRAY['int', 'str', 'bool', 'float'], 2),
('Входной тест: основы Python', 'Что выведет print(2 + 3 * 2)?', 2, ARRAY['10', '12', '8', '7'], 3),
('Входной тест: основы Python', 'Какой тип данных представляет True?', 3, ARRAY['str', 'int', 'float', 'bool'], 4),
('Входной тест: основы Python', 'Какой функцией читают ввод пользователя?', 4, ARRAY['input()', 'read()', 'scan()', 'get()'], 1),
('Проверка: основы Python', 'Что вернёт type(7)?', 0, ARRAY['<class ''str''>', '<class ''int''>', '<class ''float''>', '<class ''bool''>'], 2),
('Проверка: основы Python', 'Чему равно выражение 17 % 5?', 1, ARRAY['3', '3.4', '2', '5'], 3),
('Проверка: основы Python', 'Что вернёт int(''12'')?', 2, ARRAY['Строку ''12''', 'Число 12.0', 'Ошибка в любом случае', 'Число 12'], 4),
('Проверка: основы Python', 'Какое имя переменной допустимо?', 3, ARRAY['user_score', '2score', 'user-score', 'class'], 1),
('Проверка: условия и циклы', 'Какой оператор проверяет равенство двух значений?', 0, ARRAY['=', '==', '!=', '<='], 2),
('Проверка: условия и циклы', 'Сколько раз выполнится for i in range(4)?', 1, ARRAY['3 раза', '5 раз', '4 раза', 'Бесконечно'], 3),
('Проверка: условия и циклы', 'Что делает else в конструкции if/else?', 2, ARRAY['Повторяет цикл', 'Объявляет функцию', 'Завершает программу', 'Выполняет ветку, когда условие if ложно'], 4),
('Проверка: условия и циклы', 'Что нужно изменить, чтобы while не стал бесконечным?', 3, ARRAY['Обеспечить изменение условия цикла', 'Заменить все переменные строками', 'Удалить отступы', 'Добавить ещё один while'], 1),
('Проверка: коллекции и функции', 'С какого индекса начинается список в Python?', 0, ARRAY['С единицы', 'С нуля', 'С минус единицы', 'Индексирования нет'], 2),
('Проверка: коллекции и функции', 'Как добавить элемент в конец списка items?', 1, ARRAY['items.add(value)', 'items.push(value)', 'items.append(value)', 'append(items, value)'], 3),
('Проверка: коллекции и функции', 'Какое ключевое слово объявляет функцию?', 2, ARRAY['func', 'function', 'lambda def', 'def'], 4),
('Проверка: коллекции и функции', 'Что делает return внутри функции?', 3, ARRAY['Возвращает значение вызывающему коду', 'Печатает значение на экран', 'Повторяет функцию', 'Создаёт список'], 1),
('Итоговый тест: основы Python', 'Какой тип данных у значения 3.5?', 0, ARRAY['int', 'float', 'str', 'bool'], 2),
('Итоговый тест: основы Python', 'Что выведет print(''Py'' + ''thon'')?', 1, ARRAY['Py thon', 'Ошибка сложения строк', 'Python', 'thonPy'], 3),
('Итоговый тест: основы Python', 'Чему равно 10 // 3?', 2, ARRAY['1', '3.33', '4', '3'], 4),
('Итоговый тест: основы Python', 'Как проверить, что x меньше или равен 10?', 3, ARRAY['x <= 10', 'x =< 10', 'x == 10', 'x < 10 ='], 1),
('Итоговый тест: основы Python', 'Какой блок выполняется, если условие if ложно и есть else?', 4, ARRAY['elif', 'else', 'for', 'def'], 2),
('Итоговый тест: основы Python', 'Что выведет for i in range(2): print(i)?', 5, ARRAY['1, затем 2', '0, 1, 2', '0, затем 1', 'Только 2'], 3),
('Итоговый тест: основы Python', 'Как получить значение по ключу ''name'' из словаря user?', 6, ARRAY['user.name()', 'user(name)', 'user[0]', 'user[''name'']'], 4),
('Итоговый тест: основы Python', 'Какое значение вернёт функция, если в ней нет return?', 7, ARRAY['None', '0', 'False', 'Пустая строка'], 1);

INSERT INTO test_questions (test_id, question, position)
SELECT t.id, b.question, b.question_position
FROM python_question_bank b
JOIN course_pages p ON p.title = b.page_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');

INSERT INTO test_answers (question_id, text, is_correct, position)
SELECT q.id, option_data.option_text,
       option_data.option_position = b.correct_position,
       option_data.option_position - 1
FROM python_question_bank b
JOIN course_pages p ON p.title = b.page_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
JOIN test_questions q ON q.test_id = t.id AND q.question = b.question
CROSS JOIN LATERAL unnest(b.options) WITH ORDINALITY AS option_data(option_text, option_position)
WHERE c.title IN ('Python: основы программирования', 'Python с ИИ: основы программирования');
