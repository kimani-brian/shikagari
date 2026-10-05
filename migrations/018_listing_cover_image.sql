-- 018_listing_cover_image.sql
-- Sellers and dealers can choose which photo represents the listing in search
-- results and cards. Previously the first uploaded image was always used.

ALTER TABLE IF EXISTS listings
    ADD COLUMN IF NOT EXISTS cover_image TEXT;

-- Existing listings keep showing their first photo as the cover.
UPDATE listings
SET cover_image = images[1]
WHERE cover_image IS NULL
  AND deleted_at IS NULL
  AND array_length(images, 1) > 0;