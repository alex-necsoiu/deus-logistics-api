# PostgreSQL Schema Integrity & Design

## Overview

This document describes the production-grade PostgreSQL schema for the DEUS Logistics API with comprehensive integrity constraints, foreign keys, cascading deletes, and optimized indexes.

---

## Schema Architecture

### Core Tables

#### 1. **vessels** — Vessel Master Records

```sql
CREATE TABLE vessels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,              -- Business unique identifier
    capacity NUMERIC NOT NULL CHECK (capacity > 0),
    current_location TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Constraints:**
- ✅ PRIMARY KEY: `id` (UUID)
- ✅ UNIQUE: `name` (business requirement - no duplicate vessel names)
- ✅ NOT NULL: `name`, `capacity`, `current_location`
- ✅ CHECK: `capacity > 0` (positive capacity required)

**Relationships:**
- One Vessel → Many Cargoes
- ON DELETE RESTRICT (prevent vessel deletion if carrying cargo)

**Indexes:**
- PRIMARY: `id`
- UNIQUE: `name`

---

#### 2. **cargoes** — Cargo Shipment Records

```sql
CREATE TABLE cargoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,                       -- Optional
    weight NUMERIC NOT NULL CHECK (weight > 0),
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_transit', 'delivered')),
    vessel_id UUID NOT NULL REFERENCES vessels(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Constraints:**
- ✅ PRIMARY KEY: `id` (UUID)
- ✅ FOREIGN KEY: `vessel_id` → `vessels(id)` ON DELETE RESTRICT
- ✅ NOT NULL: `name`, `weight`, `status`, `vessel_id`
- ✅ CHECK: `weight > 0` (positive weight required)
- ✅ CHECK: `status IN ('pending', 'in_transit', 'delivered')` (valid statuses only)

**Status Transition Rules:**
- `pending` → `in_transit` (shipment starts)
- `in_transit` → `delivered` (shipment completes)
- No backward transitions allowed (enforced at domain layer)

**Indexes:**
- PRIMARY: `id`
- FOREIGN KEY: `vessel_id`
- COMPOSITE: `(status, created_at DESC)` — Filter by status and time
- COMPOSITE: `(vessel_id, status)` — Find cargos by vessel and status
- PARTIAL: `(vessel_id, created_at DESC) WHERE status IN ('pending', 'in_transit')` — Active cargo

---

#### 3. **tracking_entries** — Immutable Append-Only Tracking History

```sql
CREATE TABLE tracking_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cargo_id UUID NOT NULL REFERENCES cargoes(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_transit', 'delivered')),
    note TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Constraints:**
- ✅ PRIMARY KEY: `id` (UUID)
- ✅ FOREIGN KEY: `cargo_id` → `cargoes(id)` ON DELETE CASCADE
- ✅ NOT NULL: `cargo_id`, `location`, `status`, `timestamp`
- ✅ CHECK: `status IN ('pending', 'in_transit', 'delivered')`
- ✅ TRIGGER: `tracking_append_only_trigger` (prevents UPDATE/DELETE)

**Immutability:**
- No UPDATE operations allowed (exception raised)
- No DELETE operations allowed (exception raised)
- Only INSERT operations permitted (Append-only)
- Cascades delete when cargo is removed (rare - audit only)

**Indexes:**
- PRIMARY: `id`
- FOREIGN KEY: `cargo_id`
- COMPOSITE: `(cargo_id, timestamp)` — Full audit trail
- SIMPLE: `timestamp` — Time-based queries

---

#### 4. **cargo_events** — Immutable Event Log (Kafka Consumer)

```sql
CREATE TABLE cargo_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cargo_id UUID NOT NULL REFERENCES cargoes(id) ON DELETE CASCADE,
    old_status TEXT NOT NULL CHECK (old_status IN ('pending', 'in_transit', 'delivered')),
    new_status TEXT NOT NULL CHECK (new_status IN ('pending', 'in_transit', 'delivered')),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Constraints:**
- ✅ PRIMARY KEY: `id` (UUID)
- ✅ FOREIGN KEY: `cargo_id` → `cargoes(id)` ON DELETE CASCADE
- ✅ NOT NULL: `cargo_id`, `old_status`, `new_status`, `timestamp`
- ✅ CHECK: `old_status IN ('pending', 'in_transit', 'delivered')`
- ✅ CHECK: `new_status IN ('pending', 'in_transit', 'delivered')`

**Purpose:**
- Event sourcing record (written by Kafka consumer)
- Complete status transition history
- Immutable audit trail

**Indexes:**
- PRIMARY: `id`
- FOREIGN KEY: `cargo_id`
- COMPOSITE: `(cargo_id, timestamp)` — Event timeline
- SIMPLE: `timestamp` — Time-based queries

---

## Referential Integrity

### Foreign Key Relationships

```
┌─────────────────────────────┐
│ vessels                     │
│ ├── id (PK)                 │
│ ├── name (UNIQUE)           │
│ └── ...                     │
└─────────────────────────────┘
          ↑
          │ (vessel_id) ON DELETE RESTRICT
          │
┌─────────────────────────────┐
│ cargoes                     │
│ ├── id (PK)                 │
│ ├── vessel_id (FK)   ──────→│
│ ├── status                  │
│ └── ...                     │
└─────────────────────────────┘
          ↑
          │ (cargo_id) ON DELETE CASCADE
          │
    ┌─────┴──────┐
    │            │
    │            │
┌───────────────────────┐   ┌─────────────────────┐
│ tracking_entries      │   │ cargo_events        │
│ ├── id (PK)           │   │ ├── id (PK)         │
│ ├── cargo_id (FK)  ──│   │ ├── cargo_id (FK)──│
│ ├── status            │   │ ├── old_status     │
│ ├── location          │   │ ├── new_status     │
│ └── ...               │   │ └── ...            │
└───────────────────────┘   └─────────────────────┘
```

### Cascade Behavior

| From Table | Relation | To Table | ON DELETE |
|------------|----------|----------|-----------|
| vessels | - | cargoes | RESTRICT ✅ |
| cargoes | - | tracking_entries | CASCADE ⚠️ |
| cargoes | - | cargo_events | CASCADE ⚠️ |

**Explanation:**
- ✅ RESTRICT: Prevents accidental vessel deletion if cargos exist
- ⚠️ CASCADE: AUTO-DELETES tracking/events when cargo deleted (use only in emergencies - rare)

---

## Constraints Summary

### NOT NULL Constraints

| Table | Column | Purpose |
|-------|--------|---------|
| **vessels** | name, capacity, current_location | Core vessel data required |
| **cargoes** | name, weight, status, vessel_id | Core cargo data required |
| **tracking_entries** | cargo_id, location, status, timestamp | Audit trail required |
| **cargo_events** | cargo_id, old_status, new_status, timestamp | Event log required |

### CHECK Constraints

| Table | Constraint | Values |
|-------|-----------|--------|
| **vessels** | capacity > 0 | Only positive values |
| **cargoes** | weight > 0 | Only positive values |
| **cargoes** | status | 'pending', 'in_transit', 'delivered' |
| **tracking_entries** | status | 'pending', 'in_transit', 'delivered' |
| **cargo_events** | old_status | 'pending', 'in_transit', 'delivered' |
| **cargo_events** | new_status | 'pending', 'in_transit', 'delivered' |

### UNIQUE Constraints

| Table | Column | Purpose |
|-------|--------|---------|
| **vessels** | name | Business requirement: no duplicate vessel names |

---

## Indexes Strategy

### Single-Column Indexes

| Table | Column | Purpose |
|-------|--------|---------|
| **cargoes** | vessel_id | Foreign key lookup, filter by vessel |
| **cargoes** | status | Filter cargos by status |
| **tracking_entries** | cargo_id | Foreign key lookup, retrieve history |
| **tracking_entries** | timestamp | Time-range queries |
| **cargo_events** | cargo_id | Foreign key lookup, event retrieval |
| **cargo_events** | timestamp | Time-range queries |

### Composite Indexes

| Table | Columns | Purpose | Query Pattern |
|-------|---------|---------|----------------|
| **cargoes** | (status, created_at DESC) | Filter and sort by status and time | Find recent pending cargos |
| **cargoes** | (vessel_id, status) | Filter by vessel and status | Find all pending cargos on vessel |
| **tracking_entries** | (cargo_id, timestamp) | Full audit trail retrieval | Get complete tracking history |
| **cargo_events** | (cargo_id, timestamp) | Event timeline queries | Get status transitions timeline |

### Partial Indexes

| Table | Columns | Condition | Purpose |
|-------|---------|-----------|---------|
| **cargoes** | (vessel_id, created_at DESC) | WHERE status IN ('pending', 'in_transit') | Fast active cargo lookup (smaller index) |

**Why Partial Index:**
- Active cargos (pending/in_transit) are queried much more frequently
- Delivered cargos are rarely queried yet must be preserved
- Partial index keeps active-cargo queries fast with smaller memory footprint

---

## Data Quality Guarantees

### Orphan Prevention

| Scenario | Prevention |
|----------|-----------|
| Cargo without vessel | FOREIGN KEY constraint + NOT NULL |
| Tracking entry without cargo | FOREIGN KEY constraint + NOT NULL |
| Event record without cargo | FOREIGN KEY constraint + NOT NULL |
| Invalid cargo status | CHECK constraint validates allowed values |
| Invalid tracking status | CHECK constraint validates allowed values |
| Negative capacity/weight | CHECK constraint (> 0) |
| Missing core fields | NOT NULL constraint |

### Migration Validation

The up-migration runs data integrity checks BEFORE applying constraints:

```sql
-- Detect orphaned records
SELECT orphaned_cargo WHERE vessel_id IS NULL
SELECT orphaned_tracking WHERE cargo_id IS NULL
SELECT orphaned_events WHERE cargo_id IS NULL

-- Raise exception if ANY found (prevents migration)
-- Ensures data is clean before constraints applied
```

---

## Production Readiness Checklist

### Schema Validation
- ✅ All tables have PRIMARY KEY (UUID)
- ✅ All foreign keys have explicit ON DELETE behavior
- ✅ All critical fields have NOT NULL constraints
- ✅ All enumerated fields have CHECK constraints
- ✅ Business rules have UNIQUE constraints (vessel names)
- ✅ No orphan records allowed (enforced by FKs)

### Performance Optimization
- ✅ Composite indexes for common query patterns
- ✅ Partial indexes for frequently-queried subsets
- ✅ Foreign key columns indexed (auto for PKs)
- ✅ Time-based queries optimized (timestamp indexes)

### Data Integrity
- ✅ Immutable append-only tables enforced by triggers
- ✅ Cascade deletes carefully controlled (RESTRICT for safety)
- ✅ Pre-migration data validation checks
- ✅ Comprehensive constraint documentation

### Observability
- ✅ Table comments document purpose and relationships
- ✅ Column comments document constraints and usage
- ✅ Constraint comments explain referential integrity
- ✅ Trigger comments explain immutability rules

---

## Migration Path

### Step 1: Review Current Schema
```bash
# Examine existing tables
psql -d deus_logistics -c "\d vessels"
psql -d deus_logistics -c "\d cargoes"
psql -d deus_logistics -c "\d tracking_entries"
psql -d deus_logistics -c "\d cargo_events"
```

### Step 2: Apply Migration
```bash
# Using migration tool
flyway migrate

# Or manually
psql -f internal/postgres/migrations/000003_enhance_schema_integrity.up.sql
```

### Step 3: Verify Constraints
```sql
-- Check constraints exist
SELECT constraint_name, constraint_type 
FROM information_schema.table_constraints 
WHERE table_name = 'cargoes';

-- Check indexes created
SELECT indexname FROM pg_indexes WHERE tablename = 'cargoes';
```

### Step 4: Test Data Quality
```sql
-- Verify no orphaned records
SELECT COUNT(*) FROM cargoes WHERE vessel_id NOT IN (SELECT id FROM vessels);
SELECT COUNT(*) FROM tracking_entries WHERE cargo_id NOT IN (SELECT id FROM cargoes);

-- Should both return 0
```

---

## Maintenance

### Monitoring

**Check constraint violations:**
```sql
SELECT * FROM cargoes WHERE weight <= 0;  -- Should be empty
SELECT * FROM vessels WHERE capacity <= 0;  -- Should be empty
SELECT * FROM cargoes WHERE status NOT IN ('pending', 'in_transit', 'delivered');  -- Should be empty
```

**Find orphaned records:**
```sql
-- These should always be empty if constraints work
SELECT * FROM cargoes c WHERE NOT EXISTS (SELECT 1 FROM vessels v WHERE v.id = c.vessel_id);
SELECT * FROM tracking_entries t WHERE NOT EXISTS (SELECT 1 FROM cargoes c WHERE c.id = t.cargo_id);
```

### Index Maintenance

**Rebuild bloated indexes (if needed):**
```sql
REINDEX INDEX idx_cargoes_status_created;
REINDEX INDEX idx_cargoes_active;
```

**Check index sizes:**
```sql
SELECT schemaname, tablename, indexname, pg_size_pretty(pg_relation_size(indexrelid))
FROM pg_stat_user_indexes
WHERE tablename IN ('cargoes', 'tracking_entries', 'cargo_events')
ORDER BY pg_relation_size(indexrelid) DESC;
```

---

## Conclusion

This production-grade schema provides:

✅ **Referential Integrity** — Foreign keys prevent orphan records  
✅ **Data Quality** — CHECK constraints enforce valid values  
✅ **Immutability** — Triggers prevent accidental modifications  
✅ **Performance** — Strategic indexes optimize common queries  
✅ **Compliance** — Audit trails captured and preserved  
✅ **Observability** — Comprehensive documentation and constraints  

**No orphaned records possible. Schema is production-ready.**
