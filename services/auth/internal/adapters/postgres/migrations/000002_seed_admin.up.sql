-- Passwords cannot be safely configured through a static SQL seed.
-- The auth startup bootstrap configures admin@mail.ru from ADMIN_PASSWORD.
-- Keep pgcrypto available for the following student seed migration.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
