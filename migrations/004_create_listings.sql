CREATE TABLE IF NOT EXISTS listings (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_type  VARCHAR(20)    NOT NULL CHECK (seller_type IN ('dealer', 'private')),
    title        VARCHAR(255)   NOT NULL,
    description  TEXT,
    price_kes    NUMERIC(15, 2) NOT NULL,
    location     VARCHAR(100)   NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'active'
                 CHECK (status IN ('active', 'inactive', 'sold')),
    make         VARCHAR(100)   NOT NULL,
    model        VARCHAR(100)   NOT NULL,
    year         INT            NOT NULL,
    mileage      INT            NOT NULL,           -- kilometres
    fuel_type    VARCHAR(20)    NOT NULL
                 CHECK (fuel_type IN ('petrol', 'diesel', 'hybrid', 'electric')),
    transmission VARCHAR(20)    NOT NULL
                 CHECK (transmission IN ('automatic', 'manual')),
    color        VARCHAR(50),
    images       TEXT[],                            -- array of image URLs
    view_count   INT            NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

-- Indexes to support search & filtering
CREATE INDEX IF NOT EXISTS idx_listings_user_id     ON listings (user_id);
CREATE INDEX IF NOT EXISTS idx_listings_status      ON listings (status);
CREATE INDEX IF NOT EXISTS idx_listings_location    ON listings (location);
CREATE INDEX IF NOT EXISTS idx_listings_make_model  ON listings (make, model);
CREATE INDEX IF NOT EXISTS idx_listings_price       ON listings (price_kes);
CREATE INDEX IF NOT EXISTS idx_listings_year        ON listings (year);

-- Full-text search index over title and description
CREATE INDEX IF NOT EXISTS idx_listings_fts
    ON listings USING GIN (to_tsvector('english', title || ' ' || COALESCE(description, '')));