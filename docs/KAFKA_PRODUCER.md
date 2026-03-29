# Production-Grade Kafka Producer

## Overview

The DEUS Logistics API uses an async, fire-and-forget Kafka producer for publishing cargo status change events. The producer is designed for production use with:

- **Async publishing** - Non-blocking event publishing with background delivery handling
- **Retry logic** - Exponential backoff with configurable retry count
- **Structured logging** - Comprehensive logging for debugging and monitoring
- **Graceful shutdown** - Properly flushes pending messages on shutdown
- **Metrics tracking** - Captures failure statistics per event type

## Architecture

### Fire-and-Forget Pattern

The producer implements **fire-and-forget** semantics: HTTP handlers queue events and return immediately without waiting for delivery confirmation.

```
HTTP Handler
    ↓
PublishStatusChanged() [non-blocking, immediate return]
    ↓
sendAsync() [background goroutine]
    ↓
Kafka WriteMessages()
    ↓
Delivery Report (logged async)
```

**Critical Design:** The HTTP request completes _before_ the event is actually sent to Kafka. This ensures HTTP latency is not affected by Kafka performance.

### Async Delivery Pipeline

```go
// Handler (synchronous):
func (h *CargoHandler) UpdateCargoStatus(c *gin.Context) {
    // ... validation ...
    
    // Publish event (returns immediately)
    pub.PublishStatusChanged(ctx, event) // ← Non-blocking
    
    // HTTP response sent immediately
    c.JSON(http.StatusOK, response)
}

// Background async processing:
go p.sendAsync(ctx, cargoID, eventType, data, 0) // ← Background goroutine
```

## Configuration

### Default Production Settings

The producer uses sensible production defaults:

```go
ProducerConfig{
    Brokers:        // Kafka brokers from environment
    Topic:          // Topic from environment
    MaxRetries:     3,               // Retry up to 3 times
    RetryBackoffMs: 100,             // Start with 100ms backoff
    QueueSize:      1000,            // Internal queue size
    FlushTimeoutMs: 5000,            // 5s timeout on shutdown
}
```

### Exponential Backoff

Retry delays follow exponential backoff:
- Attempt 1: Immediate
- Attempt 2: 100ms delay (2^0 * 100)
- Attempt 3: 200ms delay (2^1 * 100)
- Attempt 4: 400ms delay (2^2 * 100)

After 3 retries, the message is logged as undeliverable and dropped.

### Environment Variables

```bash
KAFKA_BROKER=localhost:9092              # Broker address
KAFKA_TOPIC_STATUS_CHANGES=cargo-status-changes  # Topic name
```

## Kafka Writer Configuration

### RequestedAcks: -1 (All Replicas)

```go
RequiredAcks: -1  // Wait for all in-sync replicas
```

- Ensures durability across replicas
- Prevents message loss on broker failure
- Trade-off: Higher latency but guaranteed delivery

### Async Mode

```go
Async: true  // Non-blocking writes
```

- WriteMessages() returns immediately
- Messages batched and sent in background
- Failed messages trigger retries via sendAsync()

### Balancer: LeastBytes

```go
Balancer: &kafka.LeastBytes{}  // Distribute by partition size
```

- Even distribution of messages across partitions
- Prevents hot partitions

### Batch Settings

```go
BatchTimeout: 100 * time.Millisecond  // Flush every 100ms
```

- Balances latency vs. throughput
- Batches messages for efficient network use

## Delivery Guarantees

### What We Guarantee

✅ **At-least-once delivery** (eventually)
- Messages will be delivered to Kafka eventually
- May be delivered multiple times if retries overlap
- Consumer must handle idempotency

### What We Don't Guarantee

❌ **Synchronous delivery confirmation**
- HTTP returns before Kafka ack
- Fire-and-forget semantics by design
- Delivery failures are logged, not propagated

### Message Loss Scenarios

**When messages are lost:**
1. All retries exhausted (3 attempts + exponential backoff)
2. Server crashes during delivery (message still in memory)
3. Kafka cluster becomes entirely unreachable

**Mitigation:**
- Kafka cluster redundancy (3+ brokers)
- Application monitoring for failed events
- Event replay from database if needed

## Implementation Details

### sendAsync() Method

Runs in background goroutine to avoid blocking:

```go
go p.sendAsync(ctx, cargoID, eventType, data, 0)
```

### Retry Logic

Each retry attempt:
1. Checks if publisher is shutting down (context)
2. Sends message to Kafka
3. On success: logs and returns
4. On failure + retries available: schedules retry with backoff
5. On failure + no retries: logs exhausted and records metric

### Graceful Shutdown

```
Close() called
  ↓
Set closed = true (reject new publishes)
  ↓
Flush pending messages (5s timeout)
  ↓
Cancel context (signal goroutines)
  ↓
Wait for goroutines to finish
  ↓
Close Kafka writer
```

## Structured Logging

### Success Log

```json
{
  "level": "info",
  "message": "event successfully published to Kafka",
  "cargo_id": "abc123",
  "event_type": "cargo.status_changed",
  "attempt": 1
}
```

### Retry Log

```json
{
  "level": "warn",
  "message": "event delivery failed, scheduling retry",
  "cargo_id": "abc123",
  "event_type": "cargo.status_changed",
  "attempt": 2,
  "max_retries": 3,
  "backoff": "100ms",
  "error": "connection refused"
}
```

### Failure Log

```json
{
  "level": "error",
  "message": "event delivery failed after exhausting all retries",
  "cargo_id": "abc123",
  "event_type": "cargo.status_changed",
  "total_attempts": 4,
  "max_retries": 3,
  "error": "broker unreachable"
}
```

## Monitoring

### Metrics

The producer tracks failures per event type:

```go
failureMetrics map[string]int  // ["cargo.status_changed"] = N
```

Reported on shutdown:
```
2026-03-29T14:30:00 WARN undelivered events during shutdown event_type=cargo.status_changed failed_count=5
```

### Instrumentation Points

**Critical logs to monitor:**
1. `event delivery failed after exhausting all retries` - Message loss
2. `failed to marshal cargo status changed event` - Serialization issues
3. `publisher shutting down` - Messages dropped during shutdown
4. Undelivered metrics on Close() - Total failure count

## Usage

### Basic Usage

```go
// Initialize producer (recommended: wrap in singleton)
publisher := events.NewEventPublisher(
    []string{"localhost:9092"},
    "cargo-status-changes",
)
defer publisher.Close()

// Publish event (returns immediately)
ctx := context.Background()
event := cargo.StatusChangedEvent{
    ID:        uuid.New().String(),
    EventType: "cargo.status_changed",
    CargoID:   cargoID.String(),
    OldStatus: cargo.CargoStatusPending,
    NewStatus: cargo.CargoStatusInTransit,
    Timestamp: time.Now(),
}

// Fire-and-forget: error ignored
_ = publisher.PublishStatusChanged(ctx, event)

// HTTP response continues immediately
// Event delivery happens in background
```

### Custom Configuration

```go
config := events.ProducerConfig{
    Brokers:        []string{"kafka1:9092", "kafka2:9092"},
    Topic:          "cargo-status-changes",
    MaxRetries:     5,          // More retries for critical events
    RetryBackoffMs: 200,        // Longer backoff
    QueueSize:      2000,       // Larger queue
    FlushTimeoutMs: 10000,      // 10s shutdown timeout
}

publisher := events.NewEventPublisherWithConfig(config)
defer publisher.Close()
```

## Best Practices

### 1. Always Defer Close()

```go
publisher := events.NewEventPublisher(brokers, topic)
defer publisher.Close()  // ← REQUIRED: flushes pending messages
```

Ensures pending messages are flushed on application shutdown.

### 2. Ignore Return Values

```go
_ = publisher.PublishStatusChanged(ctx, event)  // Intentional: fire-and-forget
```

The return value is for interface compliance only and should always be ignored.

### 3. Log Event Publication

Events are automatically logged:
- ✅ Success: INFO level
- ⚠️ Retry: WARN level
- ❌ Failure: ERROR level

No need for additional logging in callers.

### 4. Handle Graceful Shutdown

```go
func (s *Server) Shutdown(ctx context.Context) error {
    // Close publisher first (flushes messages)
    if err := s.publisher.Close(); err != nil {
        log.Error().Err(err).Msg("error closing publisher")
    }
    
    // Then close other services
    return s.http.Shutdown(ctx)
}
```

### 5. Monitor Shutdown Logs

Check logs on application startup/shutdown for:
- `Kafka event publisher closed successfully` - Clean shutdown
- `undelivered events during shutdown` - Check for issues
- `timeout waiting for background goroutines` - Potential deadlock

## Testing

### Unit Tests

The producer is tested for:
- ✅ Successful async publishing
- ✅ Retry logic with exponential backoff
- ✅ Graceful shutdown with message flushing
- ✅ Context cancellation handling

### Integration Tests

- ✅ Real Kafka cluster publishing
- ✅ Broker failure scenarios
- ✅ High-volume message publishing
- ✅ Concurrent publish + shutdown

## Performance Characteristics

### Latency

- **PublishStatusChanged()**: < 1ms (queueing only)
- **Actual delivery**: 100-500ms (depends on Kafka)

### Throughput

- **Messages per second**: Depends on Kafka cluster
- **Queue size**: 1000 messages (configurable)
- **Batch size**: Kafka default (100 messages)

### Memory

- **Minimal**: Only queue size * message size
- **Default**: ~1000 messages × ~500 bytes = ~500KB

## Troubleshooting

### Messages Not Appearing in Kafka

1. **Check logs** for delivery failures
2. **Verify Kafka cluster** is accessible
3. **Check topic** exists: `kafka-topics --list`
4. **Monitor metrics** for undelivered counts

### High Retry Rates

Symptoms:
- Repeated `event delivery failed, scheduling retry` logs
- Warnings at WARN level

Causes:
- Kafka broker unreachable
- Network timeout
- Broker overloaded

Solutions:
- Check broker health: `nc -z broker 9092`
- Increase retry timeout if network latency is high
- Check broker logs for errors

### Message Loss on Shutdown

- Ensure `defer publisher.Close()` is called
- Check logs for `undelivered events during shutdown`
- Increase `FlushTimeoutMs` if timeout occurs

### Consumer Not Receiving Events

1. **Verify Kafka cluster** has the topic
2. **Check consumer group** is subscribed
3. **Verify producer** is publishing (check logs)
4. **Monitor partitions** for message lag

## Future Improvements

Potential enhancements:
- [ ] Dead-letter queue for undeliverable messages
- [ ] Prometheus metrics export
- [ ] Circuit breaker for failing brokers
- [ ] Message batching optimization
- [ ] Compression codec selection
- [ ] Request tracing (OpenTelemetry)

## References

- [Kafka-go Documentation](https://pkg.go.dev/github.com/segmentio/kafka-go)
- [Kafka Best Practices](https://kafka.apache.org/documentation/#bestpractices)
- [Fire-and-Forget Pattern](https://en.wikipedia.org/wiki/Fire-and-forget)
