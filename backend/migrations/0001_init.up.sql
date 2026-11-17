CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS endpoints (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          TEXT NOT NULL UNIQUE,
    name          TEXT,
    is_enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    request_count BIGINT NOT NULL DEFAULT 0,
    last_request_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_endpoints_created_at ON endpoints (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_endpoints_expires_at ON endpoints (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id   UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    method        TEXT NOT NULL,
    path          TEXT NOT NULL DEFAULT '',
    raw_query     TEXT,
    query_params  JSONB NOT NULL DEFAULT '{}'::jsonb,
    headers       JSONB NOT NULL DEFAULT '{}'::jsonb,
    content_type  TEXT,
    body          BYTEA,
    body_size     BIGINT NOT NULL DEFAULT 0,
    body_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    source_ip     INET,
    received_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_endpoint_received ON events (endpoint_id, received_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_received_at ON events (received_at);
