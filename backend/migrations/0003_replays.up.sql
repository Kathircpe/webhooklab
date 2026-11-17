CREATE TABLE IF NOT EXISTS replays (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id              UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    destination_url       TEXT NOT NULL,
    method                TEXT NOT NULL,
    request_headers       JSONB NOT NULL DEFAULT '{}'::jsonb,
    request_body          BYTEA,
    response_status       INTEGER,
    response_headers      JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_body         BYTEA,
    response_duration_ms  BIGINT NOT NULL DEFAULT 0,
    error                 TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_replays_event_created ON replays (event_id, created_at DESC);
