-- Rename internal worker-managed incidents to outage_events.
-- These are purely for uptime math and alert triggering, not user-visible incidents.
ALTER TABLE incidents RENAME TO outage_events;
ALTER INDEX idx_incidents_monitor RENAME TO idx_outage_events_monitor;
ALTER INDEX idx_incidents_open    RENAME TO idx_outage_events_open;
