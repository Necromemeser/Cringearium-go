-- Remove all seeded Python and humanities courses; dependent content cascades automatically.
DELETE FROM courses
WHERE title IN ('Python: основы программирования', 'Python с ИИ: основы программирования', 'Русский язык: подготовка к ЕГЭ', 'Русский язык с ИИ: подготовка к ЕГЭ', 'Основы психологии', 'Основы психологии с ИИ', 'Философия без занудства', 'Философия без занудства с ИИ', 'Мемология: история и теория мемов', 'Мемология с ИИ: история и теория мемов');

ALTER TABLE tests ALTER COLUMN passing_score SET DEFAULT 50;
