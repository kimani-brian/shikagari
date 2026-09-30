-- 013_add_engine_size_to_listings.sql
-- Optional engine size spec (e.g. 3.0L) shown on listing cards.

ALTER TABLE IF EXISTS listings
    ADD COLUMN IF NOT EXISTS engine_size VARCHAR(20);
