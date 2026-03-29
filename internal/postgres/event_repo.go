package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// EventRepository implements cargo.EventRepository using sqlc generated code.
// This table is APPEND-ONLY — written by Kafka consumer.
type EventRepository struct {
	pool *pgxpool.Pool
}

// NewEventRepository creates a new event repository.
func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

// Store persists a cargo event record (APPEND-ONLY).
func (r *EventRepository) Store(ctx context.Context, event cargo.CargoEvent) error {
	const query = `
		INSERT INTO cargo_events (cargo_id, old_status, new_status, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query,
		event.CargoID,
		event.OldStatus.String(),
		event.NewStatus.String(),
		event.Timestamp,
	)

	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", event.CargoID.String()).
			Str("old_status", event.OldStatus.String()).
			Str("new_status", event.NewStatus.String()).
			Msg("failed to store cargo event")
		return fmt.Errorf("storing cargo event: %w", err)
	}

	zerolog.Ctx(ctx).Debug().
		Str("cargo_id", event.CargoID.String()).
		Str("old_status", event.OldStatus.String()).
		Str("new_status", event.NewStatus.String()).
		Msg("cargo event stored in database")
	return nil
}

// ListByCargoID retrieves all events for a cargo in chronological order.
func (r *EventRepository) ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*cargo.CargoEvent, error) {
	const query = `
		SELECT id, cargo_id, old_status, new_status, timestamp
		FROM cargo_events
		WHERE cargo_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.pool.Query(ctx, query, cargoID)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", cargoID.String()).
			Msg("failed to query cargo events")
		return nil, fmt.Errorf("listing cargo events: %w", err)
	}
	defer rows.Close()

	var events []*cargo.CargoEvent
	for rows.Next() {
		var e cargo.CargoEvent
		var oldStatus, newStatus string

		if err := rows.Scan(
			&e.ID,
			&e.CargoID,
			&oldStatus,
			&newStatus,
			&e.Timestamp,
		); err != nil {
			zerolog.Ctx(ctx).Error().
				Err(err).
				Str("cargo_id", cargoID.String()).
				Msg("failed to scan cargo event row")
			return nil, fmt.Errorf("scanning cargo event: %w", err)
		}

		e.OldStatus = cargo.CargoStatus(oldStatus)
		e.NewStatus = cargo.CargoStatus(newStatus)
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", cargoID.String()).
			Msg("rows iteration error for cargo events")
		return nil, fmt.Errorf("rows error: %w", err)
	}

	zerolog.Ctx(ctx).Debug().
		Int("count", len(events)).
		Str("cargo_id", cargoID.String()).
		Msg("cargo events retrieved from database")
	return events, nil
}
