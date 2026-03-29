-- name: StoreCargoEvent :one
INSERT INTO cargo_events (cargo_id, old_status, new_status, timestamp)
VALUES ($1, $2, $3, NOW())
RETURNING id, cargo_id, old_status, new_status, timestamp;

-- name: ListCargoEventsByCargoID :many
SELECT id, cargo_id, old_status, new_status, timestamp
FROM cargo_events
WHERE cargo_id = $1
ORDER BY timestamp ASC;

-- name: CountCargoEventsByCargoID :one
SELECT COUNT(*) as count
FROM cargo_events
WHERE cargo_id = $1;
