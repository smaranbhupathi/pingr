-- Component status for public status page display
-- Separate from internal up/down — reflects what the operator wants to communicate.
ALTER TABLE monitors
    ADD COLUMN component_status TEXT NOT NULL DEFAULT 'operational',
    ADD COLUMN description       TEXT,
    ADD COLUMN component_id      UUID;

CREATE TABLE components (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    sort_order  INT  NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE monitors
    ADD CONSTRAINT fk_monitors_component
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE SET NULL;
