-- Backfill distributor schema for deployments that applied an older 058 migration variant.

ALTER TABLE IF EXISTS distributor_profiles
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS distributor_offers
    ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE IF EXISTS distributor_offers
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS sell_price_cny_cents BIGINT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS memo TEXT NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS redeemed_at TIMESTAMPTZ;

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS revoked_by_admin BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE IF EXISTS distributor_orders
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE IF EXISTS distributor_wallet_ledger
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF to_regclass('public.distributor_orders') IS NOT NULL THEN
        UPDATE distributor_orders
        SET sell_price_cny_cents = cost_cny_cents
        WHERE sell_price_cny_cents IS NULL OR sell_price_cny_cents = 0;

        CREATE INDEX IF NOT EXISTS idx_distributor_orders_user_issued_at
            ON distributor_orders(distributor_user_id, issued_at DESC);
        CREATE INDEX IF NOT EXISTS idx_distributor_orders_status
            ON distributor_orders(status);
    END IF;

    IF to_regclass('public.distributor_wallet_ledger') IS NOT NULL THEN
        CREATE INDEX IF NOT EXISTS idx_distributor_wallet_ledger_user_created_at
            ON distributor_wallet_ledger(distributor_user_id, created_at DESC);
        CREATE INDEX IF NOT EXISTS idx_distributor_wallet_ledger_type
            ON distributor_wallet_ledger(type);
    END IF;
END $$;
