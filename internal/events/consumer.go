package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

type EventConsumer struct {
	reader *kafka.Reader
	repo   cargo.EventRepository
	done   chan struct{}
}

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

func (c *EventConsumer) Stop() error {
	close(c.done)
	return c.reader.Close()
}
