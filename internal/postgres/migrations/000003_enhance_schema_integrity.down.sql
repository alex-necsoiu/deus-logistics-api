-- Revert schema integrity enhancements
-- This migration removes constraints, indexes, and validations added for production-grade schema

-- 1. Drop composite indexes
DROP INDEX IF EXISTS idx_cargoes_status_created;
DROP INDEX IF EXISTS idx_cargoes_vessel_status;
DROP INDEX IF EXISTS idx_tracking_entries_cargo_timestamp;
DROP INDEX IF EXISTS idx_cargo_events_cargo_timestamp;
DROP INDEX IF EXISTS idx_cargoes_active;

-- 2. Drop UNIQUE constraint on vessel names
ALTER TABLE vessels
  DROP CONSTRAINT IF EXISTS unique_vessel_name;

-- 3. Drop CHECK constraints on cargo_events
ALTER TABLE cargo_events
  DROP CONSTRAINT IF EXISTS check_cargo_events_old_status;

ALTER TABLE cargo_events
  DROP CONSTRAINT IF EXISTS check_cargo_events_new_status;

-- 4. Drop CHECK constraints on tracking_entries
ALTER TABLE tracking_entries
  DROP CONSTRAINT IF EXISTS check_tracking_entries_status;

-- 5. Drop CHECK constraints on vessels
ALTER TABLE vessels
  DROP CONSTRAINT IF EXISTS check_vessel_capacity_positive;

-- 6. Drop CHECK constraints on cargoes
ALTER TABLE cargoes
  DROP CONSTRAINT IF EXISTS check_cargo_weight_positive;

-- 7. Remove comments (optional, PostgreSQL allows null comments for cleanup)
COMMENT ON CONSTRAINT fk_cargoes_vessel_id ON cargoes IS NULL;
COMMENT ON CONSTRAINT fk_tracking_entries_cargo_id ON tracking_entries IS NULL;
COMMENT ON CONSTRAINT fk_cargo_events_cargo_id ON cargo_events IS NULL;
COMMENT ON TABLE cargoes IS NULL;
COMMENT ON TABLE vessels IS NULL;
COMMENT ON TABLE tracking_entries IS NULL;
COMMENT ON TABLE cargo_events IS NULL;
COMMENT ON COLUMN cargoes.status IS NULL;
COMMENT ON COLUMN cargoes.vessel_id IS NULL;
COMMENT ON COLUMN tracking_entries.status IS NULL;
COMMENT ON COLUMN cargo_events.old_status IS NULL;
COMMENT ON COLUMN cargo_events.new_status IS NULL;
