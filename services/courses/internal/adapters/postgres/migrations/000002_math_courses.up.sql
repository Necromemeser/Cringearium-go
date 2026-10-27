-- Seed both final math-course variants directly, without a temporary starter course.
INSERT INTO courses (title, theme, description, price, image_id, author_id, status)
VALUES (
    'Математика: курс с ИИ',
    'Математика',
    'Полный курс по линейным, квадратным и тригонометрическим уравнениям: теория, примеры, обычные тесты и адаптивная практика с ИИ.',
    0, NULL, NULL, 'published'
);

INSERT INTO courses (title, theme, description, price, image_id, author_id, status)
SELECT 'Математика: обычные тесты', 'Математика',
       'Тот же курс по линейным, квадратным и тригонометрическим уравнениям, но вся практика проходит по фиксированным тестам без адаптивной генерации.',
       0, NULL, NULL, 'published'
WHERE NOT EXISTS (SELECT 1 FROM courses WHERE title = 'Математика: обычные тесты');

INSERT INTO course_sections (course_id, title, description, position)
SELECT c.id, s.title, s.description, s.position
FROM courses c
CROSS JOIN (VALUES
    ('Линейные уравнения', 'Преобразование уравнений и поиск неизвестного.', 0),
    ('Квадратные уравнения', 'Неполные уравнения, дискриминант, формула корней и теорема Виета.', 1),
    ('Тригонометрические уравнения', 'Базовые уравнения, единичная окружность и периодичность решений.', 2)
) AS s(title, description, position)
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO course_pages (section_id, title, type, content, position)
SELECT s.id, p.title,
       CASE WHEN p.page_type = 'adaptive' AND c.title = 'Математика: курс с ИИ'
            THEN 'ai_test'::page_type
            WHEN p.page_type = 'adaptive' THEN 'test'::page_type
            ELSE p.page_type::page_type END,
       p.content, p.position
FROM courses c
JOIN course_sections s ON s.course_id = c.id
CROSS JOIN (VALUES
('Линейные уравнения','Что такое линейное уравнение','theory',
'# Линейные уравнения

Линейное уравнение с одной переменной приводится к виду **ax + b = 0**, где a и b — числа, а a не равно нулю.

Решить уравнение — значит найти все значения x, при которых левая и правая части равны.

**Пример:** 3x − 7 = 11. Прибавим 7 к обеим частям: 3x = 18. Разделим на 3: x = 6.

Проверка: 3 · 6 − 7 = 11. Равенство верно.',0),
('Линейные уравнения','Преобразования и проверка решения','theory',
'# Как решать уравнения

1. Раскройте скобки по распределительному закону.
2. Приведите подобные слагаемые.
3. Выполните одинаковые действия с обеими частями.
4. Разделите на коэффициент при x, если он не равен нулю.
5. Подставьте ответ в исходное уравнение.

**Пример:** 2(x + 3) = 14 → 2x + 6 = 14 → 2x = 8 → x = 4.

Если получается 0x = 5, решений нет. Если получается 0x = 0, подходит любое действительное число.',1),
('Линейные уравнения','Проверка: линейные уравнения','test',NULL,2),
('Линейные уравнения','Практика с обратной связью: линейные уравнения','adaptive',NULL,3),
('Квадратные уравнения','Неполные квадратные уравнения','theory',
'# Неполные квадратные уравнения

Квадратное уравнение имеет вид **ax² + bx + c = 0**, где a не равно нулю.

Если b или c равны нулю, уравнение называют неполным.

**Пример:** x² − 9 = 0 → x² = 9 → x = −3 или x = 3.

Для уравнения x² + 6x = 0 вынесем x за скобки: x(x + 6) = 0. Значит, x = 0 или x = −6.',0),
('Квадратные уравнения','Дискриминант','theory',
'# Дискриминант

Для уравнения ax² + bx + c = 0 дискриминант вычисляется по формуле:

**D = b² − 4ac**

- D > 0 — два различных действительных корня;
- D = 0 — один действительный корень;
- D < 0 — действительных корней нет.

**Пример:** x² − 5x + 6 = 0. Здесь a = 1, b = −5, c = 6, поэтому D = 25 − 24 = 1.',1),
('Квадратные уравнения','Формула корней','theory',
'# Формула корней

Если D ≥ 0, корни квадратного уравнения находятся по формуле:

**x₁,₂ = (−b ± √D) / 2a**

Для x² − 5x + 6 = 0 дискриминант равен 1: x₁ = (5 + 1) / 2 = 3, x₂ = (5 − 1) / 2 = 2.

Подстановка корней в исходное уравнение помогает обнаружить вычислительную ошибку.',2),
('Квадратные уравнения','Теорема Виета','theory',
'# Теорема Виета

Для приведённого уравнения x² + px + q = 0 сумма корней равна −p, а произведение — q.

Для x² − 5x + 6 = 0 нужны два числа, сумма которых равна 5, а произведение — 6. Это числа 2 и 3.

Теорема Виета помогает проверять ответы и иногда находить корни без вычисления дискриминанта.',3),
('Квадратные уравнения','Проверка: квадратные уравнения','test',NULL,4),
('Квадратные уравнения','Практика: квадратные уравнения','test',NULL,5),
('Квадратные уравнения','Практика с обратной связью: квадратные уравнения','adaptive',NULL,6),
('Тригонометрические уравнения','Тригонометрия и единичная окружность','theory',
'# Тригонометрические уравнения

Решением тригонометрического уравнения может быть целое семейство значений. В формулах n обозначает любое целое число, а углы измеряются в радианах.

- sin x и cos x имеют период 2π;
- tg x имеет период π.

Единичная окружность помогает понять, где синус и косинус принимают заданные значения.',0),
('Тригонометрические уравнения','Базовые тригонометрические уравнения','theory',
'# Базовые решения

- sin x = 0 → x = πn;
- cos x = 1 → x = 2πn;
- tg x = 0 → x = πn,

где n — любое целое число.

У синуса и косинуса решения повторяются через 2π, у тангенса — через π. При решении на заданном промежутке из общего семейства нужно выбрать подходящие значения.',1),
('Тригонометрические уравнения','Решения на промежутке','theory',
'# Решения на заданном промежутке

Сначала найдите общее решение, затем подберите целые значения n, которые попадают в нужный промежуток.

**Пример:** cos x = 1 на отрезке [0; 2π]. Общее решение: x = 2πn. На указанном отрезке подходят x = 0 и x = 2π.

Проверяйте, включены ли границы промежутка: это зависит от записи скобок.',2),
('Тригонометрические уравнения','Проверка: тригонометрические уравнения','test',NULL,3),
('Тригонометрические уравнения','Практика с обратной связью I: тригонометрия','adaptive',NULL,4),
('Тригонометрические уравнения','Практика с обратной связью II: тригонометрия','adaptive',NULL,5)
) AS p(section_title, title, page_type, content, position)
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
  AND s.title = p.section_title;

INSERT INTO tests (page_id, passing_score)
SELECT p.id, 70
FROM course_pages p
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE p.type = 'test'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

CREATE TEMP TABLE math_question_bank (
    page_title TEXT NOT NULL,
    question TEXT NOT NULL,
    question_position INTEGER NOT NULL,
    options TEXT[] NOT NULL,
    correct_position INTEGER NOT NULL
) ON COMMIT DROP;

INSERT INTO math_question_bank VALUES
('Проверка: линейные уравнения','Решите уравнение: 3x − 7 = 11.',0,ARRAY['x = 6','x = 4','x = −6','x = 18'],1),
('Проверка: линейные уравнения','Решите уравнение: 2(x + 3) = 14.',1,ARRAY['x = 4','x = 7','x = 5','x = 10'],1),
('Проверка: линейные уравнения','Что получится из уравнения 0x = 5?',2,ARRAY['Решений нет','Любое число','Только x = 0','Только x = 5'],1),
('Практика с обратной связью: линейные уравнения','Решите уравнение: 5x + 4 = 2x + 19.',0,ARRAY['x = 5','x = 3','x = 15','x = −5'],1),
('Практика с обратной связью: линейные уравнения','Решите уравнение: (x − 2) / 3 = 4.',1,ARRAY['x = 14','x = 10','x = 6','x = 12'],1),
('Практика с обратной связью: линейные уравнения','Какое уравнение имеет бесконечно много решений?',2,ARRAY['2x + 6 = 2(x + 3)','2x + 6 = 2x + 5','3x = 12','x + 1 = 0'],1),
('Проверка: квадратные уравнения','В уравнении 2x² − 3x + 7 = 0 чему равен коэффициент a?',0,ARRAY['2','−3','7','0'],1),
('Проверка: квадратные уравнения','Чему равен дискриминант уравнения x² − 5x + 6 = 0?',1,ARRAY['1','25','−1','0'],1),
('Проверка: квадратные уравнения','Что можно сказать о действительных корнях, если D < 0?',2,ARRAY['Действительных корней нет','Всегда два корня','Всегда один корень','Оба корня равны нулю'],1),
('Практика: квадратные уравнения','Найдите корни уравнения x² − 5x + 6 = 0.',0,ARRAY['x = 2 и x = 3','x = −2 и x = −3','x = 1 и x = 6','Корней нет'],1),
('Практика: квадратные уравнения','Найдите корни уравнения 2x² − 8 = 0.',1,ARRAY['x = −2 и x = 2','x = 0 и x = 4','x = −4 и x = 4','Только x = 2'],1),
('Практика: квадратные уравнения','Чему равна сумма корней уравнения x² − 5x + 6 = 0?',2,ARRAY['5','6','−5','1'],1),
('Практика с обратной связью: квадратные уравнения','Найдите корни уравнения x² − 9 = 0.',0,ARRAY['x = −3 и x = 3','x = 0 и x = 9','Только x = 3','Корней нет'],1),
('Практика с обратной связью: квадратные уравнения','Решите уравнение x² + 6x + 9 = 0.',1,ARRAY['x = −3','x = 3 и x = −3','x = 9','Действительных корней нет'],1),
('Практика с обратной связью: квадратные уравнения','Какие корни уравнения x² − 7x + 12 = 0?',2,ARRAY['x = 3 и x = 4','x = −3 и x = −4','x = 2 и x = 6','Корней нет'],1),
('Проверка: тригонометрические уравнения','Как записывается общее решение уравнения sin x = 0?',0,ARRAY['x = πn, n ∈ ℤ','x = π/2 + 2πn','x = 2πn + π/4','x = π/2 + πn'],1),
('Проверка: тригонометрические уравнения','Как записывается общее решение уравнения cos x = 1?',1,ARRAY['x = 2πn, n ∈ ℤ','x = π + 2πn','x = π/2 + πn','x = πn'],1),
('Проверка: тригонометрические уравнения','Чему равен основной период функции sin x?',2,ARRAY['2π','π','π/2','4π'],1),
('Практика с обратной связью I: тригонометрия','Как записывается общее решение уравнения sin x = 1?',0,ARRAY['x = π/2 + 2πn, n ∈ ℤ','x = πn','x = π + 2πn','x = 2πn'],1),
('Практика с обратной связью I: тригонометрия','Как записывается общее решение уравнения cos x = 0?',1,ARRAY['x = π/2 + πn, n ∈ ℤ','x = 2πn','x = πn','x = π/4 + πn'],1),
('Практика с обратной связью I: тригонометрия','Как записывается общее решение уравнения tg x = 0?',2,ARRAY['x = πn, n ∈ ℤ','x = π/2 + πn','x = 2πn + π/4','x = π/2 + 2πn'],1),
('Практика с обратной связью II: тригонометрия','Как записывается общее решение уравнения sin x = −1?',0,ARRAY['x = −π/2 + 2πn, n ∈ ℤ','x = π/2 + 2πn','x = πn','x = π + 2πn'],1),
('Практика с обратной связью II: тригонометрия','Как записывается общее решение уравнения cos x = −1?',1,ARRAY['x = π + 2πn, n ∈ ℤ','x = 2πn','x = π/2 + πn','x = πn'],1),
('Практика с обратной связью II: тригонометрия','Как записывается общее решение уравнения tg x = 1?',2,ARRAY['x = π/4 + πn, n ∈ ℤ','x = π/2 + πn','x = 2πn','x = π/4 + 2πn'],1);

INSERT INTO test_questions (test_id, question, position)
SELECT t.id, b.question, b.question_position
FROM math_question_bank b
JOIN course_pages p ON p.title = b.page_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO test_answers (question_id, text, is_correct, position)
SELECT q.id, option_data.option_text,
       option_data.option_position = b.correct_position,
       option_data.option_position - 1
FROM math_question_bank b
JOIN course_pages p ON p.title = b.page_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
JOIN test_questions q ON q.test_id = t.id AND q.question = b.question
CROSS JOIN LATERAL unnest(b.options) WITH ORDINALITY AS option_data(option_text, option_position)
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');


-- Expand trigonometry theory and add matched pre/post assessments to both study conditions.

UPDATE course_pages p
SET content = CASE p.title
    WHEN 'Тригонометрия и единичная окружность' THEN
'# Тригонометрические уравнения: с чего начать

Тригонометрическое уравнение содержит неизвестный угол под знаком sin, cos, tg или ctg. В отличие от линейного уравнения, у него часто не один ответ, а бесконечное семейство ответов: значения функций повторяются через определённый период.

## Радианы и градусы

Полный оборот равен 360° или 2π радиан, половина оборота — 180° или π радиан, четверть оборота — 90° или π/2 радиан.

| Градусы | Радианы |
|---|---|
| 0° | 0 |
| 30° | π/6 |
| 45° | π/4 |
| 60° | π/3 |
| 90° | π/2 |
| 180° | π |
| 270° | 3π/2 |
| 360° | 2π |

В формулах общего решения углы обычно записывают в радианах. Буква n обозначает любое целое число: ..., −2, −1, 0, 1, 2, ...

## Единичная окружность

Единичная окружность имеет радиус 1 и центр в начале координат. Точке, соответствующей углу x, отвечают координаты (cos x; sin x). Поэтому косинус — это горизонтальная координата, а синус — вертикальная.

Знаки функций по четвертям:
- I четверть: sin x > 0 и cos x > 0;
- II четверть: sin x > 0, cos x < 0;
- III четверть: sin x < 0 и cos x < 0;
- IV четверть: sin x < 0, cos x > 0.

Тангенс равен sin x / cos x и не определён, когда cos x = 0.

## Периодичность

sin x и cos x повторяют значения через 2π: sin(x + 2π) = sin x, cos(x + 2π) = cos x. Тангенс и котангенс повторяются через π, в точках, где они определены.

Важно: если уравнение имеет несколько решений, нельзя ограничиваться только одним углом на окружности. Нужно записать общее решение.'
    WHEN 'Базовые тригонометрические уравнения' THEN
'# Синус и косинус: как найти все углы

## Уравнение sin x = a

Синус может принимать значения только от −1 до 1. Если |a| > 1, действительных решений нет.

Если |a| ≤ 1, обозначим α = arcsin(a), где α — главное значение арксинуса. Тогда общее решение:
**x = α + 2πn** или **x = π − α + 2πn**, где n ∈ ℤ.

Частные случаи, которые полезно запомнить:
- sin x = 0 → x = πn;
- sin x = 1 → x = π/2 + 2πn;
- sin x = −1 → x = −π/2 + 2πn.

Пример: sin x = 1/2. На единичной окружности синус равен 1/2 при углах π/6 и 5π/6. Поэтому x = π/6 + 2πn или x = 5π/6 + 2πn.

## Уравнение cos x = a

Косинус также лежит в диапазоне [−1; 1]. Если |a| > 1, действительных решений нет.

Если |a| ≤ 1, пусть α = arccos(a), где α ∈ [0; π]. Тогда:
**x = ±α + 2πn**, где n ∈ ℤ.

Частные случаи:
- cos x = 1 → x = 2πn;
- cos x = 0 → x = π/2 + πn;
- cos x = −1 → x = π + 2πn.

Пример: cos x = 1/2. Углы на окружности — π/3 и 5π/3, поэтому x = ±π/3 + 2πn.

Проверяйте знак и период: у синуса два семейства решений, а запись одного угла без периода обычно не является полным ответом.'
    WHEN 'Решения на промежутке' THEN
'# Тангенс, котангенс и отбор корней

## Тангенс и котангенс

Тангенс определяется формулой tg x = sin x / cos x. Он не существует, когда cos x = 0. Его основной период равен π.

Если tg x = a, то:
**x = arctg(a) + πn**, где n ∈ ℤ.

Частные случаи:
- tg x = 0 → x = πn;
- tg x = 1 → x = π/4 + πn;
- tg x = −1 → x = −π/4 + πn.

Котангенс определяется как ctg x = cos x / sin x и не существует при sin x = 0. Для уравнения ctg x = a:
**x = arcctg(a) + πn**, если используется главное значение arcctg из интервала (0; π).

## Как решить уравнение на заданном промежутке

1. Решите уравнение и запишите все семейства корней.
2. Подставляйте целые значения n, начиная с тех, которые могут попасть в указанный промежуток.
3. Оставьте только значения внутри промежутка.
4. Проверьте границы: квадратные скобки означают, что граница включена, круглые — что не включена.
5. Не считайте один и тот же угол дважды, если он получен из двух эквивалентных записей.

Пример: cos x = 1 на [0; 2π]. Общее решение x = 2πn. На отрезке лежат x = 0 и x = 2π. Обе границы включены, поэтому подходят оба значения.

Пример: tg x = 0 на (0; π). Общее решение x = πn. Ни 0, ни π не входят в открытый интервал, поэтому решений на этом промежутке нет.'
    ELSE p.content
END
FROM course_sections s
JOIN courses c ON c.id = s.course_id
WHERE p.section_id = s.id
  AND s.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

-- Temporarily move existing trig pages to avoid position collisions.
UPDATE course_pages p
SET position = p.position + 20
FROM course_sections s
JOIN courses c ON c.id = s.course_id
WHERE p.section_id = s.id
  AND s.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO course_pages (section_id, title, type, content, position)
SELECT s.id, p.title, p.page_type::page_type, p.content, p.position
FROM courses c
JOIN course_sections s ON s.course_id = c.id
CROSS JOIN (VALUES
    ('Тригонометрические уравнения', 'Таблица значений и знаки функций', 'theory',
'# Таблица значений и знаки функций

Полезные значения на единичной окружности:

| x | sin x | cos x | tg x |
|---|---:|---:|---:|
| 0 | 0 | 1 | 0 |
| π/6 | 1/2 | √3/2 | √3/3 |
| π/4 | √2/2 | √2/2 | 1 |
| π/3 | √3/2 | 1/2 | √3 |
| π/2 | 1 | 0 | не определён |
| π | 0 | −1 | 0 |

Чтобы определить знак, сначала найдите четверть окружности. Синус соответствует вертикальной координате, косинус — горизонтальной. Тангенс имеет знак отношения синуса к косинусу.

Пример: угол 2π/3 находится во второй четверти. Поэтому sin(2π/3) положителен, cos(2π/3) отрицателен, а tg(2π/3) отрицателен.', 3),
    ('Тригонометрические уравнения', 'Основные тождества и преобразования', 'theory',
'# Основные тождества и преобразования

Главное тригонометрическое тождество:
**sin² x + cos² x = 1**.

Из него следуют соотношения:
- sin² x = 1 − cos² x;
- cos² x = 1 − sin² x;
- tg x = sin x / cos x, если cos x ≠ 0;
- 1 + tg² x = 1 / cos² x, если cos x ≠ 0.

Иногда уравнение удобнее решить после переноса всех слагаемых в одну часть, вынесения общего множителя или применения тождества.

Пример: 2sin x − 1 = 0 → sin x = 1/2. Затем используйте общее решение для синуса.

Пример: 2cos² x − 1 = 0 → cos² x = 1/2 → cos x = ±√2/2. Важно сохранить оба знака при извлечении квадратного корня.

Не делите обе части на выражение с неизвестным, пока не проверили, не может ли оно быть равно нулю: так можно потерять решения.', 4),
    ('Тригонометрические уравнения', 'Алгоритм решения: от уравнения к ответу', 'theory',
'# Алгоритм решения тригонометрических уравнений

1. Упростите уравнение, если это возможно: раскройте скобки, перенесите слагаемые или примените тождество.
2. Приведите его к виду sin x = a, cos x = a, tg x = a или к знакомому частному случаю.
3. Проверьте область допустимых значений и диапазон функции. Например, для синуса и косинуса правая часть должна принадлежать [−1; 1].
4. Запишите все семейства решений с периодом.
5. Если дан промежуток, отберите только корни, попадающие в него.
6. Подставьте найденные значения в исходное уравнение или проверьте их с помощью единичной окружности.

Типичная ошибка — записать только один угол. Например, sin x = 1/2 имеет два семейства корней за полный оборот, а не одно.

Ещё одна ошибка — перепутать периоды: у sin и cos период 2π, у tg и ctg — π.', 5),
    ('Тригонометрические уравнения', 'Входной тест: тригонометрия', 'test', NULL, 6),
    ('Тригонометрические уравнения', 'Итоговый тест: тригонометрия', 'test', NULL, 10)
) AS p(section_title, title, page_type, content, position)
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
  AND s.title = p.section_title;

UPDATE course_pages p
SET position = CASE p.title
    WHEN 'Тригонометрия и единичная окружность' THEN 0
    WHEN 'Базовые тригонометрические уравнения' THEN 1
    WHEN 'Решения на промежутке' THEN 2
    WHEN 'Таблица значений и знаки функций' THEN 3
    WHEN 'Основные тождества и преобразования' THEN 4
    WHEN 'Алгоритм решения: от уравнения к ответу' THEN 5
    WHEN 'Входной тест: тригонометрия' THEN 6
    WHEN 'Проверка: тригонометрические уравнения' THEN 7
    WHEN 'Практика с обратной связью I: тригонометрия' THEN 8
    WHEN 'Практика с обратной связью II: тригонометрия' THEN 9
    WHEN 'Итоговый тест: тригонометрия' THEN 10
    ELSE p.position
END
FROM course_sections s
JOIN courses c ON c.id = s.course_id
WHERE p.section_id = s.id
  AND s.title = 'Тригонометрические уравнения'
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO tests (page_id, passing_score)
SELECT p.id, 0
FROM course_pages p
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE p.title IN ('Входной тест: тригонометрия', 'Итоговый тест: тригонометрия')
  AND c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

CREATE TEMP TABLE trig_assessment_questions (
    test_title TEXT NOT NULL,
    question TEXT NOT NULL,
    question_position INTEGER NOT NULL,
    options TEXT[] NOT NULL,
    correct_position INTEGER NOT NULL
) ON COMMIT DROP;

INSERT INTO trig_assessment_questions VALUES
('Входной тест: тригонометрия','Сколько градусов составляют π радиан?',0,ARRAY['90°','180°','270°','360°'],2),
('Входной тест: тригонометрия','Какова координата x точки единичной окружности, соответствующей углу α?',1,ARRAY['sin α','cos α','tg α','1'],2),
('Входной тест: тригонометрия','Каков основной период функции cos x?',2,ARRAY['π','2π','π/2','4π'],2),
('Входной тест: тригонометрия','Чему равен sin 0?',3,ARRAY['0','1','−1','Не определён'],1),
('Входной тест: тригонометрия','В какой четверти sin x положителен, а cos x отрицателен?',4,ARRAY['I','II','III','IV'],2),
('Входной тест: тригонометрия','Какое значение может принимать sin x для действительного x?',5,ARRAY['−2','−1/2','3/2','2'],2),
('Входной тест: тригонометрия','Когда tg x не определён?',6,ARRAY['Когда sin x = 0','Когда cos x = 0','Когда sin x = cos x','Когда x = 0'],2),
('Входной тест: тригонометрия','Какое равенство является основным тригонометрическим тождеством?',7,ARRAY['sin x + cos x = 1','sin² x + cos² x = 1','sin x · cos x = 1','sin² x − cos² x = 1'],2),
('Итоговый тест: тригонометрия','Запишите общее решение уравнения sin x = 0.',0,ARRAY['x = πn, n ∈ ℤ','x = 2πn + π/2','x = π/2 + πn','x = 2πn'],1),
('Итоговый тест: тригонометрия','Запишите общее решение уравнения cos x = −1.',1,ARRAY['x = 2πn','x = π + 2πn, n ∈ ℤ','x = π/2 + πn','x = πn'],2),
('Итоговый тест: тригонометрия','Какое общее решение имеет уравнение sin x = 1/2?',2,ARRAY['x = π/6 + 2πn или x = 5π/6 + 2πn','x = π/6 + πn','x = π/3 + 2πn','x = 2πn'],1),
('Итоговый тест: тригонометрия','Какое общее решение имеет уравнение cos x = 0?',3,ARRAY['x = πn','x = π/2 + πn, n ∈ ℤ','x = 2πn','x = π/4 + πn'],2),
('Итоговый тест: тригонометрия','Решите tg x = 1.',4,ARRAY['x = π/4 + πn, n ∈ ℤ','x = π/4 + 2πn','x = π/2 + πn','x = 2πn'],1),
('Итоговый тест: тригонометрия','Сколько решений имеет уравнение cos x = 2 среди действительных x?',5,ARRAY['Ни одного','Одно','Два','Бесконечно много'],1),
('Итоговый тест: тригонометрия','Какие решения уравнения cos x = 1 лежат на отрезке [0; 2π]?',6,ARRAY['Только 0','Только 2π','0 и 2π','π/2 и 3π/2'],3),
('Итоговый тест: тригонометрия','Из уравнения 2cos² x − 1 = 0 следует...',7,ARRAY['cos x = 1/2','cos x = ±√2/2','cos x = 0','cos x = ±1'],2);

INSERT INTO test_questions (test_id, question, position)
SELECT t.id, b.question, b.question_position
FROM trig_assessment_questions b
JOIN course_pages p ON p.title = b.test_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');

INSERT INTO test_answers (question_id, text, is_correct, position)
SELECT q.id, option_data.option_text,
       option_data.option_position = b.correct_position,
       option_data.option_position - 1
FROM trig_assessment_questions b
JOIN course_pages p ON p.title = b.test_title AND p.type = 'test'
JOIN tests t ON t.page_id = p.id
JOIN course_sections s ON s.id = p.section_id
JOIN courses c ON c.id = s.course_id
JOIN test_questions q ON q.test_id = t.id AND q.question = b.question
CROSS JOIN LATERAL unnest(b.options) WITH ORDINALITY AS option_data(option_text, option_position)
WHERE c.title IN ('Математика: курс с ИИ', 'Математика: обычные тесты');


CREATE TABLE assessment_results (
    id BIGSERIAL PRIMARY KEY,
    test_attempt_id BIGINT NOT NULL UNIQUE REFERENCES test_attempts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    page_id BIGINT NOT NULL REFERENCES course_pages(id) ON DELETE CASCADE,
    assessment_type TEXT NOT NULL CHECK (assessment_type IN ('pretest', 'posttest')),
    score INTEGER NOT NULL CHECK (score BETWEEN 0 AND 100),
    answers JSONB NOT NULL DEFAULT '[]'::jsonb,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX assessment_results_user_course_idx
    ON assessment_results (user_id, course_id, assessment_type);

CREATE INDEX assessment_results_completed_at_idx
    ON assessment_results (completed_at DESC);


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
SET position = position + 1000
WHERE course_id IN (
    SELECT id FROM courses
    WHERE title IN ('Математика: курс с ИИ', 'Математика: обычные тесты')
);

UPDATE course_sections
SET position = position - 999
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


ALTER TABLE tests ALTER COLUMN passing_score SET DEFAULT 50;
