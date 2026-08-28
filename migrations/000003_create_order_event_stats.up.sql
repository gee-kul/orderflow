CREATE TABLE processed_events(
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_event_stats(
    event_type TEXT PRIMARY KEY,
    event_count BIGINT NOT NULL CHECK (event_count >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);