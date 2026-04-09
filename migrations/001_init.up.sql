-- Plans (free, pro, etc.) — stored in DB so limits are configurable without code changes
CREATE TABLE plans (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 TEXT NOT NULL UNIQUE,
    max_monitors         INT  NOT NULL DEFAULT 5,
    min_interval_seconds INT  NOT NULL DEFAULT 60,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default plans
INSERT INTO plans (name, max_monitors, min_interval_seconds) VALUES
    ('free', 5, 60),
    ('pro',  50, 30);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_verified   BOOLEAN NOT NULL DEFAULT FALSE,
    verify_token  TEXT NOT NULL DEFAULT '',
    reset_token   TEXT NOT NULL DEFAULT '',
    reset_expires_at TIMESTAMPTZ,
    plan_id       UUID NOT NULL REFERENCES plans(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_verify_token ON users(verify_token) WHERE verify_token != '';
CREATE INDEX idx_users_reset_token  ON users(reset_token)  WHERE reset_token  != '';

CREATE TABLE monitors (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    url               TEXT NOT NULL,
    type              TEXT NOT NULL DEFAULT 'http',
    interval_seconds  INT  NOT NULL DEFAULT 60,
    timeout_seconds   INT  NOT NULL DEFAULT 30,
    failure_threshold INT  NOT NULL DEFAULT 2,
    region            TEXT NOT NULL DEFAULT 'us-east',
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    status            TEXT NOT NULL DEFAULT 'pending',
    last_checked_at   TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitors_user_id ON monitors(user_id);
-- Index for the worker's GetDue query — critical for performance at scale
CREATE INDEX idx_monitors_due ON monitors(region, last_checked_at)
    WHERE is_active = TRUE;

-- Partitioned by month for scalability (TimescaleDB-ready, plain PG for now)
CREATE TABLE monitor_checks (
    id              UUID NOT NULL DEFAULT gen_random_uuid(),
    monitor_id      UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_up           BOOLEAN NOT NULL,
    status_code     INT,
    response_time_ms BIGINT NOT NULL DEFAULT 0,
    error_message   TEXT NOT NULL DEFAULT '',
    region          TEXT NOT NULL DEFAULT 'us-east',
    PRIMARY KEY (id, checked_at)
) PARTITION BY RANGE (checked_at);

-- Initial partition covering the first 6 months
CREATE TABLE monitor_checks_2026_h1 PARTITION OF monitor_checks
    FOR VALUES FROM ('2026-01-01') TO ('2026-07-01');

CREATE TABLE monitor_checks_2026_h2 PARTITION OF monitor_checks
    FOR VALUES FROM ('2026-07-01') TO ('2027-01-01');

CREATE TABLE monitor_checks_2027_h1 PARTITION OF monitor_checks
    FOR VALUES FROM ('2027-01-01') TO ('2027-07-01');

CREATE INDEX idx_checks_monitor_time ON monitor_checks(monitor_id, checked_at DESC);

CREATE TABLE incidents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_incidents_monitor ON incidents(monitor_id, started_at DESC);
CREATE INDEX idx_incidents_open    ON incidents(monitor_id) WHERE resolved_at IS NULL;

-- Alert channels — JSONB config means adding Slack/Discord needs zero schema changes
CREATE TABLE alert_channels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_channels_user ON alert_channels(user_id);

-- Which monitors send alerts to which channels
CREATE TABLE alert_subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id      UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    alert_channel_id UUID NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(monitor_id, alert_channel_id)
);

CREATE INDEX idx_alert_subs_monitor ON alert_subscriptions(monitor_id);
