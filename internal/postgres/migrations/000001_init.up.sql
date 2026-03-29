-- Create vessels table
CREATE TABLE IF NOT EXISTS vessels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    capacity NUMERIC NOT NULL,
    current_location TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create cargoes table
CREATE TABLE IF NOT EXISTS cargoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    weight NUMERIC NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_transit', 'delivered')),
    vessel_id UUID NOT NULL REFERENCES vessels(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create tracking_entries table (APPEND-ONLY — NEVER UPDATE OR DELETE)
CREATE TABLE IF NOT EXISTS tracking_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cargo_id UUID NOT NULL REFERENCES cargoes(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    status TEXT NOT NULL,
    note TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create cargo_events table (APPEND-ONLY — Written by Kafka consumer)
CREATE TABLE IF NOT EXISTS cargo_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cargo_id UUID NOT NULL REFERENCES cargoes(id) ON DELETE CASCADE,
    old_status TEXT NOT NULL,
    new_status TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_cargoes_vessel_id
    ON cargoes(vessel_id);

CREATE INDEX IF NOT EXISTS idx_cargoes_status
    ON cargoes(status);

CREATE INDEX IF NOT EXISTS idx_tracking_entries_cargo_id
    ON tracking_entries(cargo_id);

CREATE INDEX IF NOT EXISTS idx_tracking_entries_timestamp
    ON tracking_entries(timestamp);

CREATE INDEX IF NOT EXISTS idx_cargo_events_cargo_id
    ON cargo_events(cargo_id);

CREATE INDEX IF NOT EXISTS idx_cargo_events_timestamp
    ON cargo_events(timestamp);
