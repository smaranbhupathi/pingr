ALTER TABLE outage_events RENAME TO incidents;
ALTER INDEX idx_outage_events_monitor RENAME TO idx_incidents_monitor;
ALTER INDEX idx_outage_events_open    RENAME TO idx_incidents_open;
