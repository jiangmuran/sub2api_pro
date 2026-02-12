-- Security chat sessions table for fast session listing

CREATE TABLE IF NOT EXISTS security_chat_sessions (
    session_id VARCHAR(160) PRIMARY KEY,
    user_id BIGINT,
    api_key_id BIGINT,
    account_id BIGINT,
    group_id BIGINT,
    platform VARCHAR(32),
    model VARCHAR(100),
    message_preview TEXT,
    first_at TIMESTAMPTZ NOT NULL,
    last_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_chat_sessions_last_at
    ON security_chat_sessions (last_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_chat_sessions_user
    ON security_chat_sessions (user_id, last_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_chat_sessions_api_key
    ON security_chat_sessions (api_key_id, last_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_chat_sessions_expires_at
    ON security_chat_sessions (expires_at);
