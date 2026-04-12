ALTER TABLE monitors DROP CONSTRAINT IF EXISTS fk_monitors_component;
DROP TABLE IF EXISTS components;
ALTER TABLE monitors
    DROP COLUMN IF EXISTS component_status,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS component_id;
