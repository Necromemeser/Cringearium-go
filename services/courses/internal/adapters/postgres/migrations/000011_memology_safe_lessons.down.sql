BEGIN;
UPDATE course_pages p SET content=b.content,updated_at=b.updated_at FROM memology_v11_original_pages b WHERE p.id=b.page_id;
DROP TABLE memology_v11_original_pages;
COMMIT;
