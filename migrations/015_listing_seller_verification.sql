-- 015_listing_seller_verification.sql
-- Buyer-created listings require NTSA e-logbook + identity review before
-- they become publicly visible.

ALTER TABLE IF EXISTS listings
    ADD COLUMN IF NOT EXISTS verification_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS verification_full_name VARCHAR(150),
    ADD COLUMN IF NOT EXISTS verification_id_number VARCHAR(30),
    ADD COLUMN IF NOT EXISTS verification_elogbook  TEXT,
    ADD COLUMN IF NOT EXISTS verified_at            TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS verified_by_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS rejection_reason       TEXT;

-- status must allow the new 'pending' value
ALTER TABLE IF EXISTS listings DROP CONSTRAINT IF EXISTS listings_status_check;
ALTER TABLE IF EXISTS listings ADD CONSTRAINT listings_status_check
    CHECK (status IN ('pending', 'active', 'inactive', 'sold'));

ALTER TABLE IF EXISTS listings ADD CONSTRAINT listings_verification_status_check
    CHECK (verification_status IN ('pending', 'approved', 'rejected'));

CREATE INDEX IF NOT EXISTS idx_listings_verification_status
    ON listings (verification_status) WHERE deleted_at IS NULL;