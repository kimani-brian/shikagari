-- 010_add_address_to_private_sellers.sql
-- Private seller profiles now collect a physical address like dealer profiles.

ALTER TABLE IF EXISTS private_seller_profiles
    ADD COLUMN IF NOT EXISTS address TEXT;
