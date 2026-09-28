-- 011_fix_business_reg_no_unique.sql
-- business_reg_no is optional, and Go persists "unset" as '' (not NULL),
-- so a full UNIQUE constraint allows only one profile without a reg number.
-- Replace it with a partial unique index: real reg numbers stay unique,
-- empty values are exempt. The service layer still rejects duplicate
-- non-empty reg numbers at the application level.

ALTER TABLE IF EXISTS dealer_profiles
    DROP CONSTRAINT IF EXISTS dealer_profiles_business_reg_no_key;

DROP INDEX IF EXISTS idx_dealer_profiles_business_reg_no;

CREATE UNIQUE INDEX IF NOT EXISTS idx_dealer_profiles_business_reg_no
    ON dealer_profiles (business_reg_no)
    WHERE business_reg_no <> '';
