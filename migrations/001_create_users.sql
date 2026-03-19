-- Enable UUID generation extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name     VARCHAR(150)  NOT NULL,
    email         VARCHAR(255)  NOT NULL UNIQUE,
    phone         VARCHAR(20)   NOT NULL,
    password_hash TEXT          NOT NULL,
    role          VARCHAR(20)   NOT NULL DEFAULT 'buyer'
                  CHECK (role IN ('buyer', 'seller', 'admin')),
    is_verified   BOOLEAN       NOT NULL DEFAULT FALSE,
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ   -- soft delete
);

CREATE INDEX IF NOT EXISTS idx_users_email     ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role      ON users (role);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);