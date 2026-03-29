-- Enhance PostgreSQL schema integrity
-- This migration adds production-grade constraints, indexes, and validations

-- 1. Add NOT NULL constraints to tracking_events fields (ensure completeness)
ALTER TABLE cargo_events
  ALTER COLUMN old_status SET NOT NULL,
  ALTER COLUMN new_status SET NOT NULL;

-- 2. Add CHECK constraint on cargo_events status values
ALTER TABLE cargo_events
  ADD CONSTRAINT check_cargo_events_old_status
    CHECK (old_status IN ('pending', 'in_transit', 'delivered'));

ALTER TABLE cargo_events
  ADD CONSTRAINT check_cargo_events_new_status
    CHECK (new_status IN ('pending', 'in_transit', 'delivered'));

-- 3. Verify tracking_entries has proper NOT NULL constraints
ALTER TABLE tracking_entries
  ALTER COLUMN cargo_id SET NOT NULL,
  ALTER COLUMN location SET NOT NULL,
  ALTER COLUMN status SET NOT NULL,
  ALTER COLUMN timestamp SET NOT NULL;

-- 4. Add CHECK constraint for tracking_entries status validation
ALTER TABLE tracking_entries
  ADD CONSTRAINT check_tracking_entries_status
    CHECK (status IN ('pending', 'in_transit', 'delivered'));

-- 5. Ensure vessel table constraints are complete
ALTER TABLE vessels
  ALTER COLUMN name SET NOT NULL,
  ALTER COLUMN capacity SET NOT NULL,
  ALTER COLUMN current_location SET NOT NULL;

-- 6. Add CHECK constraint for vessel capacity (must be positive)
ALTER TABLE vessels
  ADD CONSTRAINT check_vessel_capacity_positive
    CHECK (capacity > 0);

-- 7. Add CHECK constraint for cargo weight (must be positive)
ALTER TABLE cargoes
  ALTER COLUMN weight SET NOT NULL;

ALTER TABLE cargoes
  ADD CONSTRAINT check_cargo_weight_positive
    CHECK (weight > 0);

-- 8. Add composite index for common query patterns
-- Composite index: (status, created_at) for filtering and time-range queries
CREATE INDEX IF NOT EXISTS idx_cargoes_status_created
  ON cargoes(status, created_at DESC);

-- 9. Add composite index for vessel tracking queries
-- Composite index: (vessel_id, status) for finding cargos by vessel and status
CREATE INDEX IF NOT EXISTS idx_cargoes_vessel_status
  ON cargoes(vessel_id, status);

-- 10. Add composite index for tracking history analysis
-- Composite index: (cargo_id, timestamp) for retrieving full audit trail
CREATE INDEX IF NOT EXISTS idx_tracking_entries_cargo_timestamp
  ON tracking_entries(cargo_id, timestamp);

-- 11. Add composite index for event history analysis
-- Composite index: (cargo_id, timestamp) for event timeline queries
CREATE INDEX IF NOT EXISTS idx_cargo_events_cargo_timestamp
  ON cargo_events(cargo_id, timestamp);

-- 12. Add partial index for active cargos (optimization)
-- Non-delivered cargos are queried more frequently
CREATE INDEX IF NOT EXISTS idx_cargoes_active
  ON cargoes(vessel_id, created_at DESC)
  WHERE status IN ('pending', 'in_transit');

-- 13. Add UNIQUE constraint on vessel names (business logic)
-- Preventing duplicate vessel names improves data quality
ALTER TABLE vessels
  ADD CONSTRAINT unique_vessel_name
    UNIQUE (name);

-- 14. Add REFERENCES constraint comment documentation
COMMENT ON CONSTRAINT fk_cargoes_vessel_id ON cargoes
  IS 'Foreign key to vessels table. ON DELETE RESTRICT prevents vessel deletion if cargos exist.';

-- 15. Document cascade behavior for tracking entries
COMMENT ON CONSTRAINT fk_tracking_entries_cargo_id ON tracking_entries
  IS 'Foreign key to cargoes table. ON DELETE CASCADE automatically removes tracking history when cargo is deleted (rare - audit only).';

COMMENT ON CONSTRAINT fk_cargo_events_cargo_id ON cargo_events
  IS 'Foreign key to cargoes table. ON DELETE CASCADE automatically removes event history when cargo is deleted (rare - audit only).';

-- 16. Add schema-level documentation
COMMENT ON TABLE cargoes IS 
  'Core cargo shipment records. Status transitions: pending → in_transit → delivered. Vessel relationship is required and protected.';

COMMENT ON TABLE vessels IS 
  'Vessel master records. Can have multiple cargos. Names must be unique for business requirements.';

COMMENT ON TABLE tracking_entries IS 
  'Immutable append-only tracking history. Cannot be updated or deleted. Cascades delete when cargo is removed (rare emergency only).';

COMMENT ON TABLE cargo_events IS 
  'Immutable append-only event log (written by Kafka consumer). Contains status transitions. Cascades delete when cargo is removed.';

-- 17. Add column-level documentation for critical fields
COMMENT ON COLUMN cargoes.status IS 
  'Current cargo status: pending (awaiting shipment), in_transit (on vessel), delivered (destination reached). CHECK constraint validates allowed values.';

COMMENT ON COLUMN cargoes.vessel_id IS 
  'Foreign key to vessels. Required. ON DELETE RESTRICT prevents vessel deletion if carrying cargo. Indexed for efficient lookup.';

COMMENT ON COLUMN tracking_entries.status IS 
  'Status snapshot at time of tracking event. Must match valid cargo statuses. Part of immutable audit trail.';

COMMENT ON COLUMN cargo_events.old_status IS 
  'Status before transition. Part of immutable event record for compliance and analysis.';

COMMENT ON COLUMN cargo_events.new_status IS 
  'Status after transition. Part of immutable event record for compliance and analysis.';

-- 18. Verify no orphaned records exist (data quality check)
-- These queries will fail if orphan records are found, preventing migration if data is corrupted
DO $$
BEGIN
  -- Check for orphaned cargo records
  IF EXISTS (
    SELECT 1 FROM cargoes c
    LEFT JOIN vessels v ON c.vessel_id = v.id
    WHERE v.id IS NULL AND c.vessel_id IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'Data integrity error: Found orphaned cargo records (missing vessel references). Run cleanup before migration.';
  END IF;

  -- Check for orphaned tracking entries
  IF EXISTS (
    SELECT 1 FROM tracking_entries t
    LEFT JOIN cargoes c ON t.cargo_id = c.id
    WHERE c.id IS NULL
  ) THEN
    RAISE EXCEPTION 'Data integrity error: Found orphaned tracking entries (missing cargo references). Run cleanup before migration.';
  END IF;

  -- Check for orphaned cargo events
  IF EXISTS (
    SELECT 1 FROM cargo_events ce
    LEFT JOIN cargoes c ON ce.cargo_id = c.id
    WHERE c.id IS NULL
  ) THEN
    RAISE EXCEPTION 'Data integrity error: Found orphaned cargo events (missing cargo references). Run cleanup before migration.';
  END IF;

  RAISE NOTICE 'Data integrity check passed: No orphaned records found.';
END $$;
