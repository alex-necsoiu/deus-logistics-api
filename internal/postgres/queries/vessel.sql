-- name: CreateVessel :one
INSERT INTO vessels (name, capacity, current_location)
VALUES ($1, $2, $3)
RETURNING id, name, capacity, current_location, created_at, updated_at;

-- name: GetVesselByID :one
SELECT id, name, capacity, current_location, created_at, updated_at
FROM vessels
WHERE id = $1;

-- name: ListVessels :many
SELECT id, name, capacity, current_location, created_at, updated_at
FROM vessels
ORDER BY created_at DESC;

-- name: UpdateVesselLocation :one
UPDATE vessels
SET current_location = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, name, capacity, current_location, created_at, updated_at;

-- name: UpdateVesselCapacity :one
UPDATE vessels
SET capacity = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, name, capacity, current_location, created_at, updated_at;
