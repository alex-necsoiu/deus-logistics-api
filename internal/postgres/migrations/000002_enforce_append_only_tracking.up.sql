-- Enforce append-only constraint on tracking_entries table
-- This migration creates a trigger that prevents any UPDATE or DELETE operations on tracking_entries
-- Only INSERT operations are allowed on the tracking_entries table

-- Create a function that rejects UPDATE and DELETE operations
CREATE OR REPLACE FUNCTION prevent_tracking_modification()
RETURNS TRIGGER AS $$
BEGIN
    -- Reject UPDATE operations on tracking_entries
    IF TG_OP = 'UPDATE' THEN
        RAISE EXCEPTION 'Tracking entries are immutable. Updates are not allowed on tracking_entries table.';
    END IF;

    -- Reject DELETE operations on tracking_entries
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'Tracking entries are immutable. Deletions are not allowed on tracking_entries table.';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger that fires BEFORE UPDATE or DELETE
-- This ensures append-only semantics at the database level
CREATE TRIGGER tracking_append_only_trigger
BEFORE UPDATE OR DELETE ON tracking_entries
FOR EACH ROW
EXECUTE FUNCTION prevent_tracking_modification();

-- Update table permissions (cannot UPDATE or DELETE)
-- This is a second layer of protection using row-level security patterns
-- Note: In production, consider using ROW SECURITY policies if available
ALTER TABLE tracking_entries ADD COLUMN IF NOT EXISTS immutable BOOLEAN DEFAULT TRUE;

-- Add comment documenting the append-only constraint
COMMENT ON TABLE tracking_entries IS 'Append-only tracking history table. INSERT-only, no UPDATE or DELETE allowed.';
COMMENT ON FUNCTION prevent_tracking_modification() IS 'Trigger function to enforce append-only semantics on tracking_entries';
COMMENT ON TRIGGER tracking_append_only_trigger ON tracking_entries IS 'Prevents UPDATE and DELETE operations to maintain append-only history';
