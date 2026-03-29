-- name: CreateTrackingEntry :one
INSERT INTO tracking_entries (cargo_id, location, status, note, timestamp)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, cargo_id, location, status, note, timestamp;

-- name: ListTrackingByCargoID :many
SELECT id, cargo_id, location, status, note, timestamp
FROM tracking_entries
WHERE cargo_id = $1
ORDER BY timestamp ASC;

-- name: CountTrackingByCargoID :one
SELECT COUNT(*) as count
FROM tracking_entries
WHERE cargo_id = $1;
