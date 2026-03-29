package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

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

// Create inserts a new vessel record into the database.
//
// Inputs:
//
//	ctx   - request context for cancellation and tracing
//	input - vessel creation details (Name, Capacity, CurrentLocation required)
//
// Returns:
//
//	*Vessel with generated UUID and database-generated timestamps on success
//	Error if DB write fails or input validation fails
//
// Side effects:
//   - DB write to vessels table with auto-generated UUID and timestamps from DEFAULT NOW()
func (r *VesselRepository) Create(ctx context.Context, input vessel.CreateVesselInput) (*vessel.Vessel, error) {
	v := &vessel.Vessel{
		ID:              uuid.New(),
		Name:            input.Name,
		Capacity:        input.Capacity,
		CurrentLocation: input.CurrentLocation,
	}

	query := `INSERT INTO vessels (id, name, capacity, current_location) 
	         VALUES ($1, $2, $3, $4)
	         RETURNING id, name, capacity, current_location, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, v.ID, v.Name, v.Capacity, v.CurrentLocation).
		Scan(&v.ID, &v.Name, &v.Capacity, &v.CurrentLocation, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("vessel_id", v.ID.String()).
			Msg("failed to insert vessel")
		return nil, fmt.Errorf("insert vessel: %w", err)
	}

	zerolog.Ctx(ctx).Debug().
		Str("vessel_id", v.ID.String()).
		Str("name", v.Name).
		Float64("capacity", v.Capacity).
		Str("location", v.CurrentLocation).
		Msg("vessel inserted into database")
	return v, nil
}

// GetByID retrieves a vessel record by its UUID.
//
// Inputs:
//
//	ctx - request context for cancellation and tracing
//	id  - UUID of the vessel (must not be nil)
//
// Returns:
//
//	*Vessel on success including all fields (name, capacity, location, timestamps)
//	vessel.ErrNotFound if vessel does not exist
//	Error if DB read fails or query syntax is invalid
//
// Side effects:
//   - DB read from vessels table
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

// List retrieves all vessel records in the database.
//
// Inputs:
//
//	ctx - request context for cancellation and tracing
//
// Returns:
//
//	[]*Vessel sorted by created_at descending on success
//	Empty slice if no vessel records exist
//	Error if DB read fails or query syntax is invalid
//
// Side effects:
//   - DB read from vessels table
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

// UpdateLocation updates a vessel's current location and returns the updated record.
//
// Inputs:
//
//	ctx      - request context for cancellation and tracing
//	id       - UUID of the vessel (must not be nil)
//	location - new location value (must not be empty)
//
// Returns:
//
//	*Vessel with updated location and current timestamp on success
//	vessel.ErrNotFound if vessel does not exist
//	Error if DB write fails or query syntax is invalid
//
// Side effects:
//   - DB update to vessels table (current_location and updated_at columns)
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

// UpdateCapacity updates a vessel's cargo capacity and returns the updated record.
//
// Inputs:
//
//	ctx      - request context for cancellation and tracing
//	id       - UUID of the vessel (must not be nil)
//	capacity - new capacity value (must be positive)
//
// Returns:
//
//	*Vessel with updated capacity and current timestamp on success
//	vessel.ErrNotFound if vessel does not exist
//	Error if DB write fails or query syntax is invalid
//
// Side effects:
//   - DB update to vessels table (capacity and updated_at columns)
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
