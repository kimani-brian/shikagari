CREATE TABLE IF NOT EXISTS favorites (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    listing_id UUID        NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- A user can only save a listing once
    CONSTRAINT uq_favorites_user_listing UNIQUE (user_id, listing_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id    ON favorites (user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_listing_id ON favorites (listing_id);