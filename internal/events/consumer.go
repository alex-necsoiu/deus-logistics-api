package events

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// EventConsumer consumes cargo status change events from Kafka and persists them to the database.
// Uses a background goroutine to process events asynchronously.
// Safe for concurrent Stop() calls via sync.Once protection.
type EventConsumer struct {
	reader   *kafka.Reader
	repo     cargo.EventRepository
	done     chan struct{}
	stopOnce sync.Once
}

// NewEventConsumer creates a new Kafka event consumer.
//
// Inputs:
//
//	brokers - list of Kafka broker addresses (comma-separated hosts:ports)
//	topic   - Kafka topic name to consume from
//	groupID - consumer group ID for coordinated consumption
//	repo    - repository for persisting cargo events to database
//
// Returns:
//
//	*EventConsumer configured to read from the specified topic
//
// Side effects:
//   - Creates a new kafka.Reader instance
func NewEventConsumer(brokers []string, topic string, groupID string, repo cargo.EventRepository) *EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		CommitInterval: time.Second,
	})
	return &EventConsumer{
		reader: reader,
		repo:   repo,
		done:   make(chan struct{}),
	}
}

// Start begins consuming messages from Kafka in a background goroutine.
// Processes StatusChangedEvent messages and persists them to the event repository.
// Must be paired with Stop() to cleanly shutdown the consumer.
//
// Inputs:
//
//	ctx - context for cancellation and logging propagation
//
// Side effects:
//   - Launches a background goroutine that reads from Kafka
//   - Persists cargo.CargoEvent to the database on each successful message
//
// Note: The consumer should be stopped by calling Stop() during graceful shutdown.
func (c *EventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-c.done:
				zerolog.Ctx(ctx).Info().Msg("kafka consumer stopped")
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to read kafka message")
					continue
				}

				// Generate request_id for this event processing
				requestID := uuid.New().String()
				eventCtx := context.WithValue(ctx, "request_id", requestID)
				eventLog := zerolog.Ctx(eventCtx)

				var event cargo.StatusChangedEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					eventLog.Error().
						Err(err).
						Bytes("message", msg.Value).
						Msg("failed to unmarshal cargo status changed event")
					continue
				}

				cargoID, err := uuid.Parse(event.CargoID)
				if err != nil {
					eventLog.Error().
						Err(err).
						Str("cargo_id", event.CargoID).
						Msg("invalid cargo ID in event")
					continue
				}

				cargoEvent := cargo.CargoEvent{
					ID:        uuid.New(),
					CargoID:   cargoID,
					OldStatus: event.OldStatus,
					NewStatus: event.NewStatus,
					Timestamp: event.Timestamp,
				}

				if err := c.repo.Store(eventCtx, cargoEvent); err != nil {
					eventLog.Error().
						Err(err).
						Str("cargo_id", cargoID.String()).
						Str("old_status", event.OldStatus.String()).
						Str("new_status", event.NewStatus.String()).
						Msg("failed to store cargo event in database")
					continue
				}

				// Log business event
				eventLog.Info().
					Str("cargo_id", cargoID.String()).
					Str("old_status", event.OldStatus.String()).
					Str("new_status", event.NewStatus.String()).
					Str("timestamp", event.Timestamp.String()).
					Msg("cargo status changed event processed successfully")
			}
		}
	}()
}

// Stop signals the consumer to stop reading messages and closes the Kafka reader.
// Safe to call multiple times via sync.Once protection (only closes channel once).
//
// Returns:
//
//	Error if the Kafka reader fails to close
//
// Side effects:
//   - Signals the consumer goroutine to stop via closed done channel
//   - Closes the underlying kafka.Reader connection
func (c *EventConsumer) Stop() error {
	c.stopOnce.Do(func() {
		close(c.done)
	})
	return c.reader.Close()
}
