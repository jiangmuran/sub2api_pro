-- Add security chat logs for admin review.

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

CREATE TABLE IF NOT EXISTS security_chat_logs (
    id BIGSERIAL PRIMARY KEY,

    session_id VARCHAR(128) NOT NULL,
    request_id VARCHAR(64),
    client_request_id VARCHAR(64),

    user_id BIGINT,
    api_key_id BIGINT,
    account_id BIGINT,
    group_id BIGINT,

    platform VARCHAR(32),
    model VARCHAR(100),
    request_path VARCHAR(256),
    stream BOOLEAN NOT NULL DEFAULT false,
    status_code INT,

    messages JSONB,
    message_preview TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

COMMENT ON TABLE security_chat_logs IS 'Admin-only chat logs for security review.';

CREATE INDEX IF NOT EXISTS idx_security_chat_logs_session
    ON security_chat_logs (session_id);

CREATE INDEX IF NOT EXISTS idx_security_chat_logs_user
    ON security_chat_logs (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_chat_logs_api_key
    ON security_chat_logs (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_chat_logs_expires_at
    ON security_chat_logs (expires_at);
