CREATE TABLE IF NOT EXISTS dealer_profiles (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    business_name    VARCHAR(200) NOT NULL,
    business_reg_no  VARCHAR(100) UNIQUE,
    location         VARCHAR(100) NOT NULL,
    address          TEXT,
    logo_url         TEXT,
    description      TEXT,
    kra_pin          VARCHAR(20),
    approval_status  VARCHAR(20)  NOT NULL DEFAULT 'pending'
                     CHECK (approval_status IN ('pending', 'approved', 'rejected')),
    approved_at      TIMESTAMPTZ,
    approved_by_id   UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_dealer_profiles_approval ON dealer_profiles (approval_status);