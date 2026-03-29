package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
)

// TrackingRepository implements tracking.Repository using sqlc generated code.
// This table is APPEND-ONLY — entries are never updated or deleted.
type TrackingRepository struct {
	pool *pgxpool.Pool
}

// NewTrackingRepository creates a new tracking repository.
func NewTrackingRepository(pool *pgxpool.Pool) *TrackingRepository {
	return &TrackingRepository{pool: pool}
}

// Create inserts a new tracking entry into the database (APPEND-ONLY).
func (r *TrackingRepository) Create(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
	const query = `
		INSERT INTO tracking_entries (cargo_id, location, status, note, timestamp)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, cargo_id, location, status, note, timestamp
	`

	var t TrackingEntry
	err := r.pool.QueryRow(ctx, query,
		input.CargoID,
		input.Location,
		input.Status,
		input.Note,
	).Scan(
		&t.ID,
		&t.CargoID,
		&t.Location,
		&t.Status,
		&t.Note,
		&t.Timestamp,
	)

	if err != nil {
		return nil, fmt.Errorf("creating tracking entry: %w", err)
	}

	return &tracking.TrackingEntry{
		ID:        t.ID,
		CargoID:   t.CargoID,
		Location:  t.Location,
		Status:    t.Status,
		Note:      t.Note,
		Timestamp: t.Timestamp,
	}, nil
}

// ListByCargoID retrieves all tracking entries for a cargo in chronological order.
// Returns empty slice if no entries found.
func (r *TrackingRepository) ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*tracking.TrackingEntry, error) {
	const query = `
		SELECT id, cargo_id, location, status, note, timestamp
		FROM tracking_entries
		WHERE cargo_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.pool.Query(ctx, query, cargoID)
	if err != nil {
		return nil, fmt.Errorf("listing tracking entries: %w", err)
	}
	defer rows.Close()

	var entries []*tracking.TrackingEntry
	for rows.Next() {
		var t TrackingEntry
		if err := rows.Scan(
			&t.ID,
			&t.CargoID,
			&t.Location,
			&t.Status,
			&t.Note,
			&t.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scanning tracking entry: %w", err)
		}
		entries = append(entries, &tracking.TrackingEntry{
			ID:        t.ID,
			CargoID:   t.CargoID,
			Location:  t.Location,
			Status:    t.Status,
			Note:      t.Note,
			Timestamp: t.Timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return entries, nil
}
