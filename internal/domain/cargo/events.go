package cargo

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event type constants.
const (
	EventTypeStatusChanged = "cargo.status_changed"
)

// CargoEvent is an immutable record of a cargo status transition.
// Written to the cargo_events table by the Kafka consumer worker.
type CargoEvent struct {
	ID        uuid.UUID
	CargoID   uuid.UUID
	OldStatus CargoStatus
	NewStatus CargoStatus
	Timestamp time.Time
}

// StatusChangedEvent is the Kafka message payload emitted on every status change.
type StatusChangedEvent struct {
	ID        string      `json:"id"`
	EventType string      `json:"event_type"`
	CargoID   string      `json:"cargo_id"`
	OldStatus CargoStatus `json:"old_status"`
	NewStatus CargoStatus `json:"new_status"`
	Timestamp time.Time   `json:"timestamp"`
}

// EventPublisher defines the contract for publishing cargo domain events.
// Implemented by internal/events/producer.go (Kafka).
// Interface keeps domain decoupled from Kafka specifics.
type EventPublisher interface {
	// PublishStatusChanged emits a cargo status change event to Kafka.
	// Must be called after a successful DB write, never before.
	// Errors are logged but MUST NOT fail the HTTP request (fire-and-forget).
	PublishStatusChanged(ctx context.Context, event StatusChangedEvent) error
}

// EventRepository defines the contract for persisting cargo events.
// Implemented by internal/postgres/event_repo.go.
type EventRepository interface {
	// Store persists a cargo event record. Append-only — never updated or deleted.
	Store(ctx context.Context, event CargoEvent) error

	// ListByCargoID retrieves all events for a cargo in chronological order.
	ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*CargoEvent, error)
}
