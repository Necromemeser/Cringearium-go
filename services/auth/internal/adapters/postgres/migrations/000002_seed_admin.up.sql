CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (
    username,
    email,
    password_hash,
    role
)
VALUES (
    'admin',
    'admin@mail.ru',
    crypt('admin123', gen_salt('bf', 12)),
    'admin'
)
ON CONFLICT (email) DO UPDATE
SET username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role;
