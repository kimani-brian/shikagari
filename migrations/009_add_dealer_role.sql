-- 009_add_dealer_role.sql
-- Splits the generic seller into private sellers ('seller') and business
-- dealerships ('dealer'). Widens the role CHECK constraint accordingly.

ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('buyer', 'seller', 'dealer', 'admin'));
