DELETE FROM users
WHERE id IN (SELECT user_id FROM seeded_student_users);

DROP TABLE seeded_student_users;
