CREATE TABLE IF NOT EXISTS user_sessions (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device        VARCHAR(120) NOT NULL,
    browser       VARCHAR(120) NOT NULL,
    location      VARCHAR(120) NOT NULL,
    ip_address    VARCHAR(45)  NOT NULL,
    user_agent    TEXT,
    last_active   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    revoked_at    TIMESTAMPTZ,
    revoked_reason VARCHAR(160),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_revoked_at ON user_sessions (revoked_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_active ON user_sessions (last_active DESC);
