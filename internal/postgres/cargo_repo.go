package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// CargoRepository implements cargo.Repository using PostgreSQL.
type CargoRepository struct {
	pool *pgxpool.Pool
}

// NewCargoRepository creates a new cargo repository with the given connection pool.
//
// Inputs:
//
//	pool - PostgreSQL connection pool (must not be nil)
//
// Returns:
//
//	*CargoRepository with initialized pool
func NewCargoRepository(pool *pgxpool.Pool) *CargoRepository {
	return &CargoRepository{pool: pool}
}

// Create inserts a new cargo record into the database.
//
// Inputs:
//
//	ctx   - request context for cancellation and tracing
//	input - cargo creation details (Name, Description, Weight, VesselID required)
//
// Returns:
//
//	*Cargo with generated UUID and CargoStatusPending on success
//	Error if DB write fails or vessel does not exist
//
// Side effects:
//   - DB write to cargoes table with auto-generated UUID and timestamps
func (r *CargoRepository) Create(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
	c := &cargo.Cargo{
		ID:          uuid.New(),
		VesselID:    input.VesselID,
		Name:        input.Name,
		Description: input.Description,
		Weight:      input.Weight,
		Status:      cargo.CargoStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	query := `INSERT INTO cargoes (id, vessel_id, name, description, weight, status, created_at, updated_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query, c.ID, c.VesselID, c.Name, c.Description, c.Weight, c.Status.String(), c.CreatedAt, c.UpdatedAt)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", c.ID.String()).
			Str("vessel_id", c.VesselID.String()).
			Msg("failed to insert cargo")
		return nil, fmt.Errorf("insert cargo: %w", err)
	}

	zerolog.Ctx(ctx).Debug().
		Str("cargo_id", c.ID.String()).
		Str("vessel_id", c.VesselID.String()).
		Str("name", c.Name).
		Msg("cargo inserted into database")
	return c, nil
}

// GetByID retrieves a cargo record by its UUID.
//
// Inputs:
//
//	ctx - request context for cancellation and tracing
//	id  - UUID of the cargo (must not be nil)
//
// Returns:
//
//	*Cargo on success including all fields (description, status, timestamps)
//	ErrNotFound if cargo does not exist
//	Error if DB read fails or query syntax is invalid
//
// Side effects:
//   - DB read from cargoes table
func (r *CargoRepository) GetByID(ctx context.Context, id uuid.UUID) (*cargo.Cargo, error) {
	c := &cargo.Cargo{}
	var statusStr string

	query := `SELECT id, vessel_id, name, description, weight, status, created_at, updated_at FROM cargoes WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.VesselID,
		&c.Name,
		&c.Description,
		&c.Weight,
		&statusStr,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zerolog.Ctx(ctx).Debug().
				Str("cargo_id", id.String()).
				Msg("cargo not found")
			return nil, cargo.ErrNotFound
		}
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", id.String()).
			Msg("failed to get cargo by id")
		return nil, fmt.Errorf("get cargo by id: %w", err)
	}

	c.Status = cargo.CargoStatus(statusStr)
	zerolog.Ctx(ctx).Debug().
		Str("cargo_id", id.String()).
		Str("status", statusStr).
		Msg("cargo retrieved from database")
	return c, nil
}

// List retrieves all cargo records in the database.
//
// Inputs:
//
//	ctx - request context for cancellation and tracing
//
// Returns:
//
//	[]*Cargo sorted by created_at descending on success
//	Empty slice if no cargo records exist
//	Error if DB read fails or query syntax is invalid
//
// Side effects:
//   - DB read from cargoes table
func (r *CargoRepository) List(ctx context.Context) ([]*cargo.Cargo, error) {
	query := `SELECT id, vessel_id, name, description, weight, status, created_at, updated_at FROM cargoes ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to query all cargoes")
		return nil, fmt.Errorf("list cargoes: %w", err)
	}
	defer rows.Close()

	var cargoes []*cargo.Cargo
	for rows.Next() {
		c := &cargo.Cargo{}
		var statusStr string

		if err := rows.Scan(&c.ID, &c.VesselID, &c.Name, &c.Description, &c.Weight, &statusStr, &c.CreatedAt, &c.UpdatedAt); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to scan cargo row")
			return nil, fmt.Errorf("scan cargo: %w", err)
		}

		c.Status = cargo.CargoStatus(statusStr)
		cargoes = append(cargoes, c)
	}
	zerolog.Ctx(ctx).Debug().Int("count", len(cargoes)).Msg("cargoes retrieved from database")
	return cargoes, rows.Err()
}

// ListByVesselID retrieves all cargo assigned to a specific vessel.
//
// Inputs:
//
//	ctx      - request context for cancellation and tracing
//	vesselID - UUID of the vessel (must not be nil)
//
// Returns:
//
//	[]*Cargo for the vessel sorted by created_at descending on success
//	Empty slice if vessel has no cargo
//	Error if DB read fails or query syntax is invalid
//
// Side effects:
//   - DB read from cargoes table filtered by vessel_id
func (r *CargoRepository) ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*cargo.Cargo, error) {
	query := `SELECT id, vessel_id, name, description, weight, status, created_at, updated_at FROM cargoes WHERE vessel_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, vesselID)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("vessel_id", vesselID.String()).
			Msg("failed to query cargoes by vessel")
		return nil, fmt.Errorf("list cargoes by vessel: %w", err)
	}
	defer rows.Close()

	var cargoes []*cargo.Cargo
	for rows.Next() {
		c := &cargo.Cargo{}
		var statusStr string

		if err := rows.Scan(&c.ID, &c.VesselID, &c.Name, &c.Description, &c.Weight, &statusStr, &c.CreatedAt, &c.UpdatedAt); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to scan cargo row")
			return nil, fmt.Errorf("scan cargo: %w", err)
		}

		c.Status = cargo.CargoStatus(statusStr)
		cargoes = append(cargoes, c)
	}
	zerolog.Ctx(ctx).Debug().
		Int("count", len(cargoes)).
		Str("vessel_id", vesselID.String()).
		Msg("vessel cargoes retrieved from database")
	return cargoes, rows.Err()
}

// UpdateStatus transitions a cargo to a new status and returns the updated record.
//
// Inputs:
//
//	ctx    - request context for cancellation and tracing
//	id     - UUID of the cargo (must not be nil)
//	status - new CargoStatus value (must be valid enum value)
//
// Returns:
//
//	*Cargo with updated status and current timestamp on success
//	ErrNotFound if cargo does not exist
//	Error if DB write fails or status is invalid
//
// Side effects:
//   - DB update to cargoes table (status and updated_at columns)
//   - Note: business logic in service layer handles tracking events and Kafka publishing
func (r *CargoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error) {
	c := &cargo.Cargo{}
	var statusStr string

	query := `UPDATE cargoes SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, vessel_id, name, description, weight, status, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, status.String(), id).Scan(
		&c.ID,
		&c.VesselID,
		&c.Name,
		&c.Description,
		&c.Weight,
		&statusStr,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zerolog.Ctx(ctx).Debug().
				Str("cargo_id", id.String()).
				Msg("cargo not found for status update")
			return nil, cargo.ErrNotFound
		}
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", id.String()).
			Str("new_status", status.String()).
			Msg("failed to update cargo status")
		return nil, fmt.Errorf("update cargo status: %w", err)
	}

	c.Status = cargo.CargoStatus(statusStr)
	zerolog.Ctx(ctx).Info().
		Str("cargo_id", id.String()).
		Str("old_status", statusStr).
		Str("new_status", statusStr).
		Msg("cargo status updated in database")
	return c, nil
}
