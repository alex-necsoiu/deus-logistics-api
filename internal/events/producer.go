package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// EventPublisher publishes events to Kafka using fire-and-forget pattern.
// Errors are logged with Zerolog for observability but never propagate to callers.
type EventPublisher struct {
	writer *kafka.Writer
	topic  string
}

// NewEventPublisher creates a new Kafka event publisher with the given brokers and topic.
//
// Inputs:
//   brokers - Kafka broker addresses (must not be empty)
//   topic   - Kafka topic name (must not be empty)
//
// Returns:
//   *EventPublisher ready to publish events
func NewEventPublisher(brokers []string, topic string) *EventPublisher {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &EventPublisher{
		writer: writer,
		topic:  topic,
	}
}

// PublishStatusChanged publishes a cargo status changed event to Kafka.
// Implements fire-and-forget pattern: errors are logged but do not fail the caller.
// Returns error for interface compliance, but caller must ignore it.
func (p *EventPublisher) PublishStatusChanged(ctx context.Context, event cargo.StatusChangedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", event.CargoID).
			Str("event_type", event.EventType).
			Msg("failed to marshal cargo status changed event")
		return fmt.Errorf("marshalling event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.CargoID),
		Value: data,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", event.CargoID).
			Str("event_type", event.EventType).
			Str("topic", p.topic).
			Msg("failed to publish cargo status changed event to Kafka")
		return fmt.Errorf("publishing event: %w", err)
	}

	zerolog.Ctx(ctx).Debug().
		Str("cargo_id", event.CargoID).
		Str("old_status", event.OldStatus.String()).
		Str("new_status", event.NewStatus.String()).
		Msg("cargo status changed event published")
	return nil
}

// Close closes the Kafka writer and releases resources.
//
// Returns:
//   Error if the close operation fails
//
// Side effects:
//   - Closes underlying Kafka connection
//   - Flushes any pending messages
func (p *EventPublisher) Close() error {
	return p.writer.Close()
}
