CREATE TABLE IF NOT EXISTS private_seller_profiles (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    national_id_no    VARCHAR(20)  NOT NULL UNIQUE,
    location          VARCHAR(100) NOT NULL,
    profile_photo_url TEXT,
    bio               TEXT,
    approval_status   VARCHAR(20)  NOT NULL DEFAULT 'pending'
                      CHECK (approval_status IN ('pending', 'approved', 'rejected')),
    approved_at       TIMESTAMPTZ,
    approved_by_id    UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_private_sellers_approval ON private_seller_profiles (approval_status);