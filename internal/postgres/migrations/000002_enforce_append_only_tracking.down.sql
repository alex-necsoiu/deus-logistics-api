-- Revert append-only enforcement on tracking_entries table
-- This migration removes the trigger and function that enforce immutability

-- Drop the trigger
DROP TRIGGER IF EXISTS tracking_append_only_trigger ON tracking_entries;

-- Drop the function
DROP FUNCTION IF EXISTS prevent_tracking_modification();

-- Remove the immutable column if it was added
-- Note: Be careful with column removal in production as it may cause data loss
ALTER TABLE tracking_entries DROP COLUMN IF EXISTS immutable;

-- Remove comments
COMMENT ON TABLE tracking_entries IS NULL;
