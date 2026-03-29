-- name: CreateCargo :one
INSERT INTO cargoes (name, description, weight, status, vessel_id)
VALUES ($1, $2, $3, 'pending', $4)
RETURNING id, name, description, weight, status, vessel_id, created_at, updated_at;

-- name: GetCargoByID :one
SELECT id, name, description, weight, status, vessel_id, created_at, updated_at
FROM cargoes
WHERE id = $1;

-- name: ListCargoes :many
SELECT id, name, description, weight, status, vessel_id, created_at, updated_at
FROM cargoes
ORDER BY created_at DESC;

-- name: ListCargosByVesselID :many
SELECT id, name, description, weight, status, vessel_id, created_at, updated_at
FROM cargoes
WHERE vessel_id = $1
ORDER BY created_at DESC;

-- name: UpdateCargoStatus :one
UPDATE cargoes
SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, name, description, weight, status, vessel_id, created_at, updated_at;
