-- 016_drop_private_seller_profiles.sql
-- Private seller accounts are retired. Buyers list directly from their
-- account, gated by NTSA e-logbook verification on the listing itself
-- (see 015_listing_seller_verification.sql).
--
-- Listings created by former private sellers keep their ownership and
-- data; the seller_type column already distinguishes "dealer" from
-- "private" on each listing.

DROP TABLE IF EXISTS private_seller_profiles;

-- Retired 'seller' accounts and their listings are removed outright.
-- Sellers who still need to trade re-register as a buyer and list again
-- through the verification flow.
DELETE FROM inquiries  WHERE listing_id IN (SELECT id FROM listings WHERE user_id IN (SELECT id FROM users WHERE role = 'seller'));
DELETE FROM favorites  WHERE listing_id IN (SELECT id FROM listings WHERE user_id IN (SELECT id FROM users WHERE role = 'seller'));
DELETE FROM listings   WHERE user_id IN (SELECT id FROM users WHERE role = 'seller');
DELETE FROM users      WHERE role = 'seller';

-- Widen the users role check to match the new three-role model.
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE IF EXISTS users ADD CONSTRAINT users_role_check
    CHECK (role IN ('buyer', 'dealer', 'admin'));