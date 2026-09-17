INSERT INTO courses (
    title,
    theme,
    description,
    price,
    image_id,
    author_id,
    status
)
VALUES (
    'Математика для СДВГ-шников',
    'Квадратные уравнения',
    'Короткий практический курс по квадратным уравнениям с теорией и небольшими проверочными тестами.',
    99000,
    NULL,
    NULL,
    'published'
);

INSERT INTO course_sections (
    course_id,
    title,
    description,
    position
)
SELECT
    id,
    'Квадратные уравнения',
    'Основные понятия, дискриминант, формула корней и практика.',
    0
FROM courses
WHERE title = 'Математика для СДВГ-шников';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'Что такое квадратное уравнение',
    'theory',
    '# Квадратное уравнение

Квадратным уравнением называется уравнение вида:

**ax² + bx + c = 0**

где **a ≠ 0**.

Коэффициенты `a`, `b` и `c` определяют конкретное квадратное уравнение.',
    0
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'Коэффициенты квадратного уравнения',
    'theory',
    '# Коэффициенты

В уравнении **ax² + bx + c = 0**:

- `a` — старший коэффициент;
- `b` — коэффициент при `x`;
- `c` — свободный член.

При этом `a` не может быть равен нулю.',
    1
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'Дискриминант',
    'theory',
    '# Дискриминант

Дискриминант квадратного уравнения вычисляется по формуле:

**D = b² - 4ac**

По значению дискриминанта можно определить количество действительных корней уравнения.',
    2
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'Проверочный тест',
    'test',
    NULL,
    3
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'Формула корней',
    'theory',
    '# Формула корней

Если дискриминант квадратного уравнения неотрицателен, его корни можно найти по формуле:

**x₁,₂ = (-b ± √D) / 2a**

Если `D > 0`, уравнение имеет два действительных корня.

Если `D = 0`, уравнение имеет один действительный корень.',
    4
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'Практический тест',
    'test',
    NULL,
    5
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO course_pages (
    section_id,
    title,
    type,
    content,
    position
)
SELECT
    id,
    'AI-тест',
    'ai_test',
    NULL,
    6
FROM course_sections
WHERE title = 'Квадратные уравнения';

INSERT INTO tests (
    page_id,
    passing_score
)
SELECT
    id,
    70
FROM course_pages
WHERE title = 'Проверочный тест';

INSERT INTO tests (
    page_id,
    passing_score
)
SELECT
    id,
    70
FROM course_pages
WHERE title = 'Практический тест';

INSERT INTO test_questions (
    test_id,
    question,
    position
)
SELECT
    id,
    'Как выглядит общий вид квадратного уравнения?',
    0
FROM tests
WHERE page_id = (
    SELECT id
    FROM course_pages
    WHERE title = 'Проверочный тест'
);

INSERT INTO test_questions (
    test_id,
    question,
    position
)
SELECT
    id,
    'Чему равен дискриминант уравнения x² - 5x + 6 = 0?',
    1
FROM tests
WHERE page_id = (
    SELECT id
    FROM course_pages
    WHERE title = 'Проверочный тест'
);

INSERT INTO test_questions (
    test_id,
    question,
    position
)
SELECT
    id,
    'Сколько действительных корней имеет квадратное уравнение при D > 0?',
    2
FROM tests
WHERE page_id = (
    SELECT id
    FROM course_pages
    WHERE title = 'Проверочный тест'
);

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'ax² + bx + c = 0',
    TRUE,
    0
FROM test_questions
WHERE question = 'Как выглядит общий вид квадратного уравнения?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'ax + b = 0',
    FALSE,
    1
FROM test_questions
WHERE question = 'Как выглядит общий вид квадратного уравнения?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'ax³ + bx² + cx = 0',
    FALSE,
    2
FROM test_questions
WHERE question = 'Как выглядит общий вид квадратного уравнения?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'a + bx² + c = 0',
    FALSE,
    3
FROM test_questions
WHERE question = 'Как выглядит общий вид квадратного уравнения?';


INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '-1',
    FALSE,
    0
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² - 5x + 6 = 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '0',
    FALSE,
    1
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² - 5x + 6 = 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '1',
    TRUE,
    2
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² - 5x + 6 = 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '25',
    FALSE,
    3
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² - 5x + 6 = 0?';


INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Один',
    FALSE,
    0
FROM test_questions
WHERE question = 'Сколько действительных корней имеет квадратное уравнение при D > 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Два',
    TRUE,
    1
FROM test_questions
WHERE question = 'Сколько действительных корней имеет квадратное уравнение при D > 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Три',
    FALSE,
    2
FROM test_questions
WHERE question = 'Сколько действительных корней имеет квадратное уравнение при D > 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Ни одного',
    FALSE,
    3
FROM test_questions
WHERE question = 'Сколько действительных корней имеет квадратное уравнение при D > 0?';

INSERT INTO test_questions (
    test_id,
    question,
    position
)
SELECT
    id,
    'Чему равен дискриминант уравнения x² + 4x + 4 = 0?',
    0
FROM tests
WHERE page_id = (
    SELECT id
    FROM course_pages
    WHERE title = 'Практический тест'
);

INSERT INTO test_questions (
    test_id,
    question,
    position
)
SELECT
    id,
    'Сколько корней имеет уравнение x² + 2x + 5 = 0 над множеством действительных чисел?',
    1
FROM tests
WHERE page_id = (
    SELECT id
    FROM course_pages
    WHERE title = 'Практический тест'
);

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '0',
    TRUE,
    0
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² + 4x + 4 = 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '4',
    FALSE,
    1
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² + 4x + 4 = 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '-4',
    FALSE,
    2
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² + 4x + 4 = 0?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    '16',
    FALSE,
    3
FROM test_questions
WHERE question = 'Чему равен дискриминант уравнения x² + 4x + 4 = 0?';


INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Два',
    FALSE,
    0
FROM test_questions
WHERE question = 'Сколько корней имеет уравнение x² + 2x + 5 = 0 над множеством действительных чисел?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Один',
    FALSE,
    1
FROM test_questions
WHERE question = 'Сколько корней имеет уравнение x² + 2x + 5 = 0 над множеством действительных чисел?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Ни одного',
    TRUE,
    2
FROM test_questions
WHERE question = 'Сколько корней имеет уравнение x² + 2x + 5 = 0 над множеством действительных чисел?';

INSERT INTO test_answers (
    question_id,
    text,
    is_correct,
    position
)
SELECT
    id,
    'Три',
    FALSE,
    3
FROM test_questions
WHERE question = 'Сколько корней имеет уравнение x² + 2x + 5 = 0 над множеством действительных чисел?';

