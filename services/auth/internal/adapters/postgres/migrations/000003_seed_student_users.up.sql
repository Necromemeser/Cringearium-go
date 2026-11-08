-- Remember only rows inserted by this migration so rollback preserves existing accounts.
CREATE TABLE seeded_student_users (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
);

WITH inserted AS (
    INSERT INTO users (username, email, password_hash, role)
    SELECT
        'user' || n,
        'user' || n || '@mail.ru',
        crypt('12345678', gen_salt('bf', 12)),
        'student'::user_role
    FROM generate_series(1, 50) AS n
    ON CONFLICT DO NOTHING
    RETURNING id
)
INSERT INTO seeded_student_users (user_id)
SELECT id FROM inserted;
