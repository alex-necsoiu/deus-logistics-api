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
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to read kafka message")
					continue
				}
				var event cargo.StatusChangedEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to unmarshal event")
					continue
				}
				cargoID, err := uuid.Parse(event.CargoID)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("invalid cargo ID in event")
					continue
				}
				cargoEvent := cargo.CargoEvent{
					ID:        uuid.New(),
					CargoID:   cargoID,
					OldStatus: event.OldStatus,
					NewStatus: event.NewStatus,
					Timestamp: event.Timestamp,
				}
				if err := c.repo.Store(ctx, cargoEvent); err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to store event")
				}
			}
		}
	}()
}

func (c *EventConsumer) Stop() error {
	close(c.done)
	return c.reader.Close()
}
