ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS body_type VARCHAR(20) NOT NULL DEFAULT 'SUV'
        CHECK (body_type IN ('SUV', 'Sedan', 'Hatchback', 'Pickup', 'Coupe', 'EV', 'Van', 'Wagon'));

-- Ensure existing rows have a valid body type value
UPDATE listings
SET body_type = 'SUV'
WHERE body_type IS NULL;

CREATE INDEX IF NOT EXISTS idx_listings_body_type ON listings (body_type);
