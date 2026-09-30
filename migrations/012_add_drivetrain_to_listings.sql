-- 012_add_drivetrain_to_listings.sql
-- Optional drivetrain spec (2WD, 4WD, AWD) shown on listing cards.

ALTER TABLE IF EXISTS listings
    ADD COLUMN IF NOT EXISTS drivetrain VARCHAR(10);
