-- Remove outbound source footnotes from learner-facing memology lessons only.

-- Original source records remain in the repository, not linked from lessons.

BEGIN;

CREATE TABLE memology_v11_original_pages (page_id BIGINT PRIMARY KEY REFERENCES course_pages(id) ON DELETE CASCADE, content TEXT, updated_at TIMESTAMPTZ NOT NULL);

INSERT INTO memology_v11_original_pages SELECT p.id,p.content,p.updated_at FROM course_pages p JOIN course_sections s ON s.id=p.section_id JOIN courses c ON c.id=s.course_id WHERE c.title IN ('Мемология: история и теория мемов','Мемология с ИИ: история и теория мемов') AND p.type='theory';

UPDATE course_pages SET content=replace(content,E'\n\n'||'[Оригинальная полоса KC Green](https://gunshowcomic.com/648); [источник кадра-шаблона](https://github.com/jacebrowning/memegen/tree/main/templates/fine). Это фрагмент настоящей работы автора, не генерация ИИ.',''),updated_at=now() WHERE id IN(SELECT page_id FROM memology_v11_original_pages);

UPDATE course_pages SET content=replace(content,E'\n\n'||'Источник изображения: Disney/Pixar, «История игрушек»; [каталог шаблона](https://github.com/jacebrowning/memegen/tree/main/templates/buzz), [пример семейства](https://imgflip.com/memetemplate/X-Everywhere). Это подлинный кадр, а не наша перерисовка.',''),updated_at=now() WHERE id IN(SELECT page_id FROM memology_v11_original_pages);

UPDATE course_pages SET content=replace(content,E'\n\n'||'[Справка о Doge и происхождении фотографии](https://knowyourmeme.com/memes/doge). Локальная копия настоящего шаблона взята из [каталога memegen](https://github.com/jacebrowning/memegen/tree/main/templates/doge); изображение не сгенерировано ИИ.',''),updated_at=now() WHERE id IN(SELECT page_id FROM memology_v11_original_pages);

UPDATE course_pages SET content=replace(content,E'\n\n'||'[История формата и автор исходной фотографии](https://knowyourmeme.com/memes/distracted-boyfriend); [источник локальной копии шаблона](https://github.com/jacebrowning/memegen/tree/main/templates/db). Использовано настоящее изображение без добавленной нами генерации или ретуши.',''),updated_at=now() WHERE id IN(SELECT page_id FROM memology_v11_original_pages);

UPDATE course_pages SET content=replace(content,E'\n\n'||'[История Drakeposting](https://knowyourmeme.com/memes/drakeposting); [локальная копия подлинного шаблона](https://github.com/jacebrowning/memegen/tree/main/templates/drake). Кадры не сгенерированы нами.',''),updated_at=now() WHERE id IN(SELECT page_id FROM memology_v11_original_pages);

COMMIT;
