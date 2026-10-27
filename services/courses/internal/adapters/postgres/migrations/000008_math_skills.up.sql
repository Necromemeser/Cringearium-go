CREATE TABLE learning_skills (
    code TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL
);
CREATE TABLE learning_skill_prerequisites (
    skill_code TEXT NOT NULL REFERENCES learning_skills(code) ON DELETE CASCADE,
    prerequisite_code TEXT NOT NULL REFERENCES learning_skills(code) ON DELETE CASCADE,
    PRIMARY KEY(skill_code, prerequisite_code),
    CHECK(skill_code <> prerequisite_code)
);
CREATE TABLE page_learning_skills (
    page_id BIGINT NOT NULL REFERENCES course_pages(id) ON DELETE CASCADE,
    skill_code TEXT NOT NULL REFERENCES learning_skills(code) ON DELETE CASCADE,
    PRIMARY KEY(page_id, skill_code)
);
INSERT INTO learning_skills(code,title,description) VALUES
('math.linear.balance','Решение простого линейного уравнения','Сохранение равенства при одинаковых действиях с обеими частями; решение ax+b=c.'),
('math.linear.transform','Преобразования линейных уравнений','Раскрытие скобок, приведение подобных слагаемых и проверка решения.'),
('math.quadratic.incomplete','Неполные квадратные уравнения','Извлечение корня и вынесение общего множителя с сохранением всех корней.'),
('math.quadratic.discriminant','Вычисление дискриминанта','Определение коэффициентов, вычисление b²−4ac и числа действительных корней.'),
('math.quadratic.roots','Корни квадратного уравнения','Применение формулы корней при D≥0 с правильными знаками и знаменателем.'),
('math.quadratic.vieta','Теорема Виета','Использование суммы и произведения корней приведённого уравнения.'),
('math.trig.circle','Единичная окружность и периоды','Значения и периоды синуса, косинуса и тангенса.'),
('math.trig.basic','Общие решения тригонометрических уравнений','Запись всех решений базовых уравнений с целочисленным параметром.'),
('math.trig.interval','Отбор решений на промежутке','Отбор корней из общего решения с учётом включения границ.');
INSERT INTO learning_skill_prerequisites VALUES
('math.linear.transform','math.linear.balance'),
('math.quadratic.incomplete','math.linear.transform'),
('math.quadratic.discriminant','math.linear.transform'),
('math.quadratic.roots','math.quadratic.discriminant'),
('math.quadratic.vieta','math.linear.balance'),
('math.trig.basic','math.trig.circle'),
('math.trig.interval','math.trig.basic');
-- Resolve authored seed pages once. Runtime selection uses stable skill codes and joins,
-- never infers a skill from question text, page order, or a generated title.
INSERT INTO page_learning_skills(page_id,skill_code)
SELECT p.id, mapping.code FROM course_pages p
JOIN course_sections s ON s.id=p.section_id JOIN courses c ON c.id=s.course_id
JOIN (VALUES
 ('Что такое линейное уравнение','math.linear.balance'),
 ('Преобразования и проверка решения','math.linear.transform'),
 ('Неполные квадратные уравнения','math.quadratic.incomplete'),
 ('Дискриминант','math.quadratic.discriminant'),
 ('Формула корней','math.quadratic.roots'),
 ('Теорема Виета','math.quadratic.vieta'),
 ('Тригонометрия и единичная окружность','math.trig.circle'),
 ('Базовые тригонометрические уравнения','math.trig.basic'),
 ('Решения на промежутке','math.trig.interval')
) mapping(title,code) ON p.title=mapping.title
WHERE p.type='theory' AND c.title IN ('Математика: курс с ИИ','Математика: обычные тесты');
