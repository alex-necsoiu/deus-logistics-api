-- Drop indexes
DROP INDEX IF EXISTS idx_cargo_events_timestamp;
DROP INDEX IF EXISTS idx_cargo_events_cargo_id;
DROP INDEX IF EXISTS idx_tracking_entries_timestamp;
DROP INDEX IF EXISTS idx_tracking_entries_cargo_id;
DROP INDEX IF EXISTS idx_cargoes_status;
DROP INDEX IF EXISTS idx_cargoes_vessel_id;

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS cargo_events;
DROP TABLE IF EXISTS tracking_entries;
DROP TABLE IF EXISTS cargoes;
DROP TABLE IF EXISTS vessels;
