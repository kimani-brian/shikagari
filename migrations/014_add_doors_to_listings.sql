-- 014_add_doors_to_listings.sql
-- Optional door count spec (0 = not specified).

ALTER TABLE IF EXISTS listings
    ADD COLUMN IF NOT EXISTS doors INTEGER NOT NULL DEFAULT 0;
