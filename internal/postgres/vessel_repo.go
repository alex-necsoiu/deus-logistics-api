package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
)

// VesselRepository implements vessel.Repository using PostgreSQL.
type VesselRepository struct {
	pool *pgxpool.Pool
}

// NewVesselRepository creates a new vessel repository.
func NewVesselRepository(pool *pgxpool.Pool) *VesselRepository {
	return &VesselRepository{pool: pool}
}

// Create inserts a new vessel record.
func (r *VesselRepository) Create(ctx context.Context, input vessel.CreateVesselInput) (*vessel.Vessel, error) {
	v := &vessel.Vessel{
		ID:              uuid.New(),
		Name:            input.Name,
		Capacity:        input.Capacity,
		CurrentLocation: input.CurrentLocation,
	}
	query := `INSERT INTO vessels (id, name, capacity, current_location, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, v.ID, v.Name, v.Capacity, v.CurrentLocation, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert vessel: %w", err)
	}
	return v, nil
}

// GetByID retrieves a vessel by ID.
func (r *VesselRepository) GetByID(ctx context.Context, id uuid.UUID) (*vessel.Vessel, error) {
	query := `SELECT id, name, capacity, current_location, created_at, updated_at FROM vessels WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	v := &vessel.Vessel{}
	err := row.Scan(&v.ID, &v.Name, &v.Capacity, &v.CurrentLocation, &v.CreatedAt, &v.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, vessel.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select vessel: %w", err)
	}
	return v, nil
}

// List retrieves all vessel records.
func (r *VesselRepository) List(ctx context.Context) ([]*vessel.Vessel, error) {
	query := `SELECT id, name, capacity, current_location, created_at, updated_at FROM vessels ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select vessels: %w", err)
	}
	defer rows.Close()
	var vessels []*vessel.Vessel
	for rows.Next() {
		v := &vessel.Vessel{}
		if err := rows.Scan(&v.ID, &v.Name, &v.Capacity, &v.CurrentLocation, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vessel: %w", err)
		}
		vessels = append(vessels, v)
	}
	return vessels, rows.Err()
}

// UpdateLocation updates a vessel's location.
func (r *VesselRepository) UpdateLocation(ctx context.Context, id uuid.UUID, location string) (*vessel.Vessel, error) {
	query := `UPDATE vessels SET current_location = $1, updated_at = NOW() WHERE id = $2 RETURNING id, name, capacity, current_location, created_at, updated_at`
	row := r.pool.QueryRow(ctx, query, location, id)
	v := &vessel.Vessel{}
	err := row.Scan(&v.ID, &v.Name, &v.Capacity, &v.CurrentLocation, &v.CreatedAt, &v.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, vessel.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update vessel: %w", err)
	}
	return v, nil
}

// UpdateCapacity updates a vessel's cargo capacity.
func (r *VesselRepository) UpdateCapacity(ctx context.Context, id uuid.UUID, capacity float64) (*vessel.Vessel, error) {
	query := `UPDATE vessels SET capacity = $1, updated_at = NOW() WHERE id = $2 RETURNING id, name, capacity, current_location, created_at, updated_at`
	row := r.pool.QueryRow(ctx, query, capacity, id)
	v := &vessel.Vessel{}
	err := row.Scan(&v.ID, &v.Name, &v.Capacity, &v.CurrentLocation, &v.CreatedAt, &v.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, vessel.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update vessel: %w", err)
	}
	return v, nil
}
