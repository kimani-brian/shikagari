CREATE TABLE IF NOT EXISTS security_events (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label     VARCHAR(120) NOT NULL,
    severity  VARCHAR(20)  NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    details   TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events (user_id);
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events (created_at DESC);
