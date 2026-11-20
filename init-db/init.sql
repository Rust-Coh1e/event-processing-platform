CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    device_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- {
--   "event_id": "fbc6b03d-13ad-4e92-b8a5-a0a8dbde1641",
--   "device_id": "sensor-001",
--   "event_type": "temperature_exceeded",
--   "occurred_at": "2026-11-16T10:00:00Z",
--   "payload": {
--     "temperature": 95.5,
--     "threshold": 90
--   }
-- }