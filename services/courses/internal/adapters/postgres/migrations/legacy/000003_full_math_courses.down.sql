DELETE FROM courses
WHERE title = 'Математика: обычные тесты';

UPDATE courses
SET title = 'Математика для СДВГ-шников',
    theme = 'Квадратные уравнения',
    description = 'Короткий практический курс по квадратным уравнениям с теорией и небольшими проверочными тестами.',
    updated_at = NOW()
WHERE title = 'Математика: курс с ИИ';
