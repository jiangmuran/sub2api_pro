-- Ensure redeem_codes has updated_at for distributor/order workflows.

ALTER TABLE IF EXISTS redeem_codes
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE redeem_codes
SET updated_at = COALESCE(updated_at, created_at, NOW())
WHERE updated_at IS NULL;
