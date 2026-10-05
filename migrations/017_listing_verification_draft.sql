-- 017_listing_verification_draft.sql
-- Two fixes on top of 015:
--   1. Dealer listings are published automatically, so existing dealer rows
--      must not sit in the admin review queue just because 015 defaulted
--      every row to 'pending'.
--   2. Buyer listings are created before their NTSA e-logbook is uploaded,
--      so 'draft' is a valid state for a listing that is still incomplete.
--      Drafts are never returned by the admin review queue.

UPDATE listings
SET verification_status = 'approved',
    verified_at         = COALESCE(verified_at, NOW())
WHERE seller_type = 'dealer'
  AND deleted_at IS NULL
  AND verification_status = 'pending';

ALTER TABLE IF EXISTS listings DROP CONSTRAINT IF EXISTS listings_verification_status_check;
ALTER TABLE IF EXISTS listings ADD CONSTRAINT listings_verification_status_check
    CHECK (verification_status IN ('draft', 'pending', 'approved', 'rejected'));