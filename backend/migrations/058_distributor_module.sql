-- Distributor module (independent from existing redeem/billing paths)

CREATE TABLE IF NOT EXISTS distributor_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    balance_cny_cents BIGINT NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_profiles_enabled ON distributor_profiles(enabled);

CREATE TABLE IF NOT EXISTS distributor_offers (
    id BIGSERIAL PRIMARY KEY,
    distributor_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    target_group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    validity_days INT NOT NULL,
    cost_cny_cents BIGINT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_offers_user_enabled ON distributor_offers(distributor_user_id, enabled);
CREATE INDEX IF NOT EXISTS idx_distributor_offers_group_id ON distributor_offers(target_group_id);

CREATE TABLE IF NOT EXISTS distributor_orders (
    id BIGSERIAL PRIMARY KEY,
    distributor_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    offer_id BIGINT NOT NULL REFERENCES distributor_offers(id) ON DELETE RESTRICT,
    redeem_code_id BIGINT NOT NULL UNIQUE REFERENCES redeem_codes(id) ON DELETE RESTRICT,
    cost_cny_cents BIGINT NOT NULL,
    sell_price_cny_cents BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'issued',
    memo TEXT NOT NULL DEFAULT '',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    redeemed_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_by_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_orders_user_issued_at ON distributor_orders(distributor_user_id, issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_distributor_orders_status ON distributor_orders(status);

CREATE TABLE IF NOT EXISTS distributor_wallet_ledger (
    id BIGSERIAL PRIMARY KEY,
    distributor_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(32) NOT NULL,
    amount_cny_cents BIGINT NOT NULL,
    balance_after_cny_cents BIGINT NOT NULL,
    operator_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    order_id BIGINT REFERENCES distributor_orders(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_wallet_ledger_user_created_at ON distributor_wallet_ledger(distributor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_distributor_wallet_ledger_type ON distributor_wallet_ledger(type);

CREATE TABLE IF NOT EXISTS distributor_settlement_checkpoints (
    id BIGSERIAL PRIMARY KEY,
    amount_cny_cents BIGINT NOT NULL,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_settlement_checkpoints_created_at ON distributor_settlement_checkpoints(created_at DESC);
