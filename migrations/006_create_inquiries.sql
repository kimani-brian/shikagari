CREATE TABLE IF NOT EXISTS inquiries (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID        NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    buyer_id   UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    seller_id  UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    message    TEXT        NOT NULL,
    reply      TEXT,
    status     VARCHAR(20) NOT NULL DEFAULT 'open'
               CHECK (status IN ('open', 'replied', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inquiries_listing_id ON inquiries (listing_id);
CREATE INDEX IF NOT EXISTS idx_inquiries_buyer_id   ON inquiries (buyer_id);
CREATE INDEX IF NOT EXISTS idx_inquiries_seller_id  ON inquiries (seller_id);
CREATE INDEX IF NOT EXISTS idx_inquiries_status     ON inquiries (status);