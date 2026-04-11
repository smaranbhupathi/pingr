-- Link auto-created incidents to the outage event that spawned them.
-- NULL for manually created incidents.
ALTER TABLE incidents ADD COLUMN outage_event_id UUID REFERENCES outage_events(id) ON DELETE SET NULL;

CREATE INDEX idx_incidents_outage_event ON incidents(outage_event_id) WHERE outage_event_id IS NOT NULL;
