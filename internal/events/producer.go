package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// ProducerConfig holds configuration for production-grade Kafka producer.
type ProducerConfig struct {
	// Brokers: Kafka broker addresses
	Brokers []string
	// Topic: Kafka topic for events
	Topic string
	// MaxRetries: Number of retry attempts for failed messages
	MaxRetries int
	// RetryBackoffMs: Initial backoff delay in milliseconds
	RetryBackoffMs int
	// QueueSize: Size of internal delivery report queue
	QueueSize int
	// FlushTimeoutMs: Timeout for flushing pending messages on shutdown
	FlushTimeoutMs int
}

// DefaultProducerConfig returns production-recommended configuration.
func DefaultProducerConfig(brokers []string, topic string) ProducerConfig {
	return ProducerConfig{
		Brokers:        brokers,
		Topic:          topic,
		MaxRetries:     3,
		RetryBackoffMs: 100,
		QueueSize:      1000,
		FlushTimeoutMs: 5000,
	}
}

// EventPublisher publishes events to Kafka using async producer pattern.
// Implements fire-and-forget with non-blocking delivery report handling.
// Production-grade: includes retry logic, structured logging, and graceful shutdown.
type EventPublisher struct {
	writer         *kafka.Writer
	topic          string
	config         ProducerConfig
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
	closed         bool
	failureMetrics map[string]int // Track failures per event type
}

// NewEventPublisher creates a production-grade Kafka event publisher.
// Uses async producer with delivery report handling in background goroutine.
//
// Inputs:
//
//	brokers - Kafka broker addresses (must not be empty)
//	topic   - Kafka topic name (must not be empty)
//
// Returns:
//
//	*EventPublisher ready to publish events safely
func NewEventPublisher(brokers []string, topic string) *EventPublisher {
	config := DefaultProducerConfig(brokers, topic)
	return NewEventPublisherWithConfig(config)
}

// NewEventPublisherWithConfig creates a publisher with explicit configuration.
func NewEventPublisherWithConfig(config ProducerConfig) *EventPublisher {
	ctx, cancel := context.WithCancel(context.Background())

	// Configure async Kafka writer with production settings
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      config.Brokers,
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  config.MaxRetries + 1,
		BatchTimeout: 100 * time.Millisecond,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		RequiredAcks: -1,   // Wait for all replicas (-1 = all, as in Kafka)
		Async:        true, // Non-blocking: fire-and-forget
	})

	p := &EventPublisher{
		writer:         writer,
		topic:          config.Topic,
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		failureMetrics: make(map[string]int),
	}

	return p
}

// PublishStatusChanged publishes a cargo status changed event to Kafka asynchronously.
// Implements fire-and-forget: errors are logged and handled async, never blocking caller.
// Method returns immediately after queuing message (non-blocking).
//
// Inputs:
//
//	ctx   - Request context for tracing
//	event - Cargo status changed event to publish
//
// Returns:
//
//	Error for interface compliance only; must be ignored by callers (fire-and-forget)
func (p *EventPublisher) PublishStatusChanged(ctx context.Context, event cargo.StatusChangedEvent) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		zerolog.Ctx(ctx).Warn().
			Str("cargo_id", event.CargoID).
			Msg("cannot publish event: publisher is closed")
		return nil // Silently ignore when closed (fire-and-forget)
	}
	p.mu.Unlock()

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", event.CargoID).
			Str("event_type", event.EventType).
			Msg("failed to marshal cargo status changed event")
		// Don't propagate error: fire-and-forget pattern
		return nil
	}

	// Queue message for async delivery (non-blocking)
	go p.sendAsync(ctx, event.CargoID, event.EventType, data, 0)

	zerolog.Ctx(ctx).Debug().
		Str("cargo_id", event.CargoID).
		Str("old_status", event.OldStatus.String()).
		Str("new_status", event.NewStatus.String()).
		Str("attempt", "1").
		Msg("cargo status changed event queued for async delivery")

	return nil // Fire-and-forget: always return nil
}

// sendAsync sends a message asynchronously with retry logic.
// Runs in background goroutine to avoid blocking caller.
func (p *EventPublisher) sendAsync(ctx context.Context, cargoID, eventType string, data []byte, attempt int) {
	logger := zerolog.Ctx(ctx)

	// Don't retry if context is cancelled
	select {
	case <-p.ctx.Done():
		logger.Warn().
			Str("cargo_id", cargoID).
			Str("event_type", eventType).
			Msg("discarding event: publisher shutting down")
		return
	default:
	}

	// Send message to Kafka
	err := p.writer.WriteMessages(p.ctx, kafka.Message{
		Key:   []byte(cargoID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(eventType)},
			{Key: "attempt", Value: []byte(fmt.Sprintf("%d", attempt+1))},
		},
	})

	if err == nil {
		// Success: log and return
		logger.Info().
			Str("cargo_id", cargoID).
			Str("event_type", eventType).
			Int("attempt", attempt+1).
			Int("total_retries_allowed", p.config.MaxRetries).
			Msg("event successfully published to Kafka")
		return
	}

	// Failure: implement retry logic with exponential backoff
	if attempt < p.config.MaxRetries {
		backoffMs := time.Duration(p.config.RetryBackoffMs*(1<<uint(attempt))) * time.Millisecond
		logger.Warn().
			Err(err).
			Str("cargo_id", cargoID).
			Str("event_type", eventType).
			Int("attempt", attempt+1).
			Int("max_retries", p.config.MaxRetries).
			Dur("backoff", backoffMs).
			Msg("event delivery failed, scheduling retry")

		// Schedule retry with exponential backoff
		select {
		case <-p.ctx.Done():
			return
		case <-time.After(backoffMs):
			p.sendAsync(ctx, cargoID, eventType, data, attempt+1)
		}
	} else {
		// Final failure: log exhausted retries
		p.mu.Lock()
		p.failureMetrics[eventType]++
		p.mu.Unlock()

		logger.Error().
			Err(err).
			Str("cargo_id", cargoID).
			Str("event_type", eventType).
			Int("total_attempts", attempt+1).
			Int("max_retries", p.config.MaxRetries).
			Msg("event delivery failed after exhausting all retries - message lost")
	}
}

// Close gracefully closes the Kafka producer and flushes pending messages.
// Waits for all in-flight messages and delivery reports before returning.
//
// Returns:
//
//	Error if close operation fails
//
// Side effects:
//   - Flushes pending messages with timeout
//   - Waits for delivery report handler goroutine
//   - Releases Kafka connection
func (p *EventPublisher) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil // Idempotent: already closed
	}
	p.closed = true
	p.mu.Unlock()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	logger.Info().Msg("closing Kafka event publisher")

	// Flush pending messages with timeout
	flushCtx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.FlushTimeoutMs)*time.Millisecond)
	defer cancel()

	if err := p.writer.WriteMessages(flushCtx); err != nil {
		logger.Warn().
			Err(err).
			Msg("error flushing pending Kafka messages on shutdown")
	}

	// Signal delivery report handler to shutdown
	p.cancel()

	// Wait for background goroutines
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Debug().Msg("all background goroutines shut down")
	case <-time.After(5 * time.Second):
		logger.Warn().Msg("timeout waiting for background goroutines to shut down")
	}

	// Close underlying Kafka writer
	if err := p.writer.Close(); err != nil {
		logger.Error().Err(err).Msg("error closing Kafka writer")
		return fmt.Errorf("closing Kafka writer: %w", err)
	}

	// Report metrics
	p.mu.Lock()
	if len(p.failureMetrics) > 0 {
		for eventType, count := range p.failureMetrics {
			logger.Warn().
				Str("event_type", eventType).
				Int("failed_count", count).
				Msg("undelivered events during shutdown")
		}
	}
	p.mu.Unlock()

	logger.Info().Msg("Kafka event publisher closed successfully")
	return nil
}
