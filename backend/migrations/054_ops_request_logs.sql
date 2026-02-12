-- Add ops_request_logs for successful request debugging (admin only).

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

CREATE TABLE IF NOT EXISTS ops_request_logs (
    id BIGSERIAL PRIMARY KEY,

    -- Correlation / identities
    request_id VARCHAR(64) NOT NULL,
    client_request_id VARCHAR(64),
    user_id BIGINT,
    api_key_id BIGINT,
    account_id BIGINT,
    group_id BIGINT,
    client_ip inet,

    -- Dimensions for filtering
    platform VARCHAR(32),

    -- Request metadata
    model VARCHAR(100),
    request_path VARCHAR(256),
    stream BOOLEAN NOT NULL DEFAULT false,
    user_agent TEXT,

    status_code INT,
    duration_ms INT,
    time_to_first_token_ms BIGINT,

    -- Sanitized payloads (admin-only)
    request_body JSONB,
    request_body_truncated BOOLEAN NOT NULL DEFAULT false,
    request_body_bytes INT,

    response_body TEXT,
    response_body_truncated BOOLEAN NOT NULL DEFAULT false,
    response_body_bytes INT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE ops_request_logs IS 'Ops request logs (success). Stores sanitized request/response for admin debugging.';

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_request_logs_request_id
    ON ops_request_logs (request_id);

CREATE INDEX IF NOT EXISTS idx_ops_request_logs_created_at
    ON ops_request_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_request_logs_platform_time
    ON ops_request_logs (platform, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_request_logs_account_time
    ON ops_request_logs (account_id, created_at DESC);
