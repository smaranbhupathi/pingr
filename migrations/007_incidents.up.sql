-- User-facing incidents (manual or auto-seeded by the worker).
-- These are what appear on the public status page timeline.
CREATE TABLE incidents (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'investigating',  -- investigating | identified | monitoring | resolved
    source      TEXT        NOT NULL DEFAULT 'manual',         -- manual | auto
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Each operator message / status change appended to an incident.
CREATE TABLE incident_updates (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID        NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    status      TEXT        NOT NULL,
    message     TEXT        NOT NULL,
    notify      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Which monitors are affected by an incident (many-to-many).
CREATE TABLE incident_affected_monitors (
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    monitor_id  UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    PRIMARY KEY (incident_id, monitor_id)
);

CREATE INDEX idx_incidents_user       ON incidents(user_id, created_at DESC);
CREATE INDEX idx_incidents_open       ON incidents(user_id) WHERE resolved_at IS NULL;
CREATE INDEX idx_incident_updates     ON incident_updates(incident_id, created_at DESC);
CREATE INDEX idx_incident_affected    ON incident_affected_monitors(monitor_id);
