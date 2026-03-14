package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pouyasadri/go-url-shortener/analytics"
	"github.com/pouyasadri/go-url-shortener/middleware"
	"github.com/pouyasadri/go-url-shortener/models"
	"github.com/pouyasadri/go-url-shortener/store"
	"github.com/redis/go-redis/v9"
)

// EventProcessor listens to Redis pub/sub and buffers events for batch writing
type EventProcessor struct {
	repo                  *analytics.Repository
	batchSize             int
	batchInterval         time.Duration
	eventBuffer           []models.RequestEvent
	errorBuffer           []models.ErrorEvent
	mu                    sync.Mutex
	redisClient           *redis.Client
	lastBatchTime         time.Time
	ctx                   context.Context
	cancel                context.CancelFunc
	wg                    sync.WaitGroup
	shutdownChan          chan struct{}
	eventsProcessed       int64
	eventsBatchWritten    int64
	metricsUpdateInterval time.Duration
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(repo *analytics.Repository, batchSize int, batchIntervalMinutes int) *EventProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventProcessor{
		repo:                  repo,
		batchSize:             batchSize,
		batchInterval:         time.Duration(batchIntervalMinutes) * time.Minute,
		eventBuffer:           make([]models.RequestEvent, 0, batchSize),
		errorBuffer:           make([]models.ErrorEvent, 0, batchSize/10), // Errors less common
		mu:                    sync.Mutex{},
		redisClient:           store.GetRedisClient(),
		lastBatchTime:         time.Now(),
		ctx:                   ctx,
		cancel:                cancel,
		shutdownChan:          make(chan struct{}),
		eventsProcessed:       0,
		eventsBatchWritten:    0,
		metricsUpdateInterval: 10 * time.Second, // Update metrics every 10 seconds
	}
}

// Start begins listening to the Redis channel and processing events
func (p *EventProcessor) Start(ctx context.Context) error {
	if p.redisClient == nil {
		return ErrRedisNotAvailable
	}

	slog.Info("Event processor starting",
		slog.Int("batch_size", p.batchSize),
		slog.Duration("batch_interval", p.batchInterval))

	p.wg.Add(1)
	go p.listenToChannel()

	// Periodic flush in case we don't reach batch size
	p.wg.Add(1)
	go p.periodicFlush()

	// Periodic metrics update
	p.wg.Add(1)
	go p.periodicMetricsUpdate()

	return nil
}

// listenToChannel subscribes to the Redis analytics:events channel
func (p *EventProcessor) listenToChannel() {
	defer p.wg.Done()

	pubsub := p.redisClient.Subscribe(p.ctx, "analytics:events")
	defer pubsub.Close()

	slog.Info("Subscribed to analytics:events channel")

	// Create channel for messages
	ch := pubsub.Channel()

	for {
		select {
		case <-p.ctx.Done():
			slog.Info("Event processor shutting down")
			// Flush remaining events
			if err := p.Flush(p.ctx); err != nil {
				slog.Error("Failed to flush final batch on shutdown", slog.String("error", err.Error()))
			}
			return
		case <-p.shutdownChan:
			slog.Info("Event processor received shutdown signal")
			// Flush remaining events
			if err := p.Flush(p.ctx); err != nil {
				slog.Error("Failed to flush final batch on shutdown", slog.String("error", err.Error()))
			}
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Parse analytics event
			var analyticsEvent middleware.AnalyticsEvent
			if err := json.Unmarshal([]byte(msg.Payload), &analyticsEvent); err != nil {
				slog.Error("Failed to unmarshal analytics event",
					slog.String("error", err.Error()),
					slog.String("payload", msg.Payload))
				continue
			}

			// Convert to RequestEvent model
			requestEvent := models.RequestEvent{
				RequestID:       analyticsEvent.RequestID,
				Timestamp:       analyticsEvent.Timestamp,
				Method:          analyticsEvent.Method,
				Path:            analyticsEvent.Path,
				StatusCode:      analyticsEvent.StatusCode,
				DurationMs:      analyticsEvent.DurationMs,
				BytesSent:       analyticsEvent.BytesSent,
				UserID:          analyticsEvent.UserID,
				APIKeyID:        analyticsEvent.APIKeyID,
				ErrorType:       analyticsEvent.ErrorType,
				CreatedShortURL: analyticsEvent.CreatedShortURL,
				ClientIP:        analyticsEvent.ClientIP,
			}

			// Add to buffer
			p.addToBuffer(requestEvent)

			// Check if we should flush
			if p.shouldFlush() {
				if err := p.Flush(p.ctx); err != nil {
					slog.Error("Failed to flush batch",
						slog.String("error", err.Error()),
						slog.Int("pending_events", len(p.eventBuffer)))
				}
			}
		}
	}
}

// addToBuffer adds an event to the processing buffer
func (p *EventProcessor) addToBuffer(event models.RequestEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.eventBuffer = append(p.eventBuffer, event)
	p.eventsProcessed++

	// If error, also add to error buffer
	if event.ErrorType != nil {
		errorEvent := models.ErrorEvent{
			RequestID:  event.RequestID,
			Timestamp:  event.Timestamp,
			ErrorType:  *event.ErrorType,
			StatusCode: event.StatusCode,
			UserID:     event.UserID,
			APIKeyID:   event.APIKeyID,
		}
		p.errorBuffer = append(p.errorBuffer, errorEvent)
	}
}

// shouldFlush determines if we should write the batch now
func (p *EventProcessor) shouldFlush() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Flush if batch size reached
	if len(p.eventBuffer) >= p.batchSize {
		return true
	}

	// Flush if interval exceeded
	if time.Since(p.lastBatchTime) >= p.batchInterval {
		return true
	}

	return false
}

// Flush writes buffered events to MongoDB
func (p *EventProcessor) Flush(ctx context.Context) error {
	p.mu.Lock()

	if len(p.eventBuffer) == 0 && len(p.errorBuffer) == 0 {
		p.mu.Unlock()
		return nil
	}

	eventsCopy := make([]models.RequestEvent, len(p.eventBuffer))
	copy(eventsCopy, p.eventBuffer)
	p.eventBuffer = p.eventBuffer[:0]

	errorsCopy := make([]models.ErrorEvent, len(p.errorBuffer))
	copy(errorsCopy, p.errorBuffer)
	p.errorBuffer = p.errorBuffer[:0]

	p.lastBatchTime = time.Now()
	p.mu.Unlock()

	// Write events to MongoDB
	if len(eventsCopy) > 0 {
		if err := p.repo.SaveRequestEventsBatch(ctx, eventsCopy); err != nil {
			slog.Error("Failed to save request events batch",
				slog.String("error", err.Error()),
				slog.Int("count", len(eventsCopy)))
			return err
		}
		p.eventsBatchWritten += int64(len(eventsCopy))
	}

	// Write errors to MongoDB (sequentially to maintain order for debugging)
	for _, errEvent := range errorsCopy {
		if err := p.repo.SaveErrorEvent(ctx, errEvent); err != nil {
			slog.Error("Failed to save error event",
				slog.String("error", err.Error()),
				slog.String("request_id", errEvent.RequestID))
			// Continue processing other errors
		}
	}

	if len(eventsCopy) > 0 {
		slog.Info("Events batch written to MongoDB",
			slog.Int("events", len(eventsCopy)),
			slog.Int("errors", len(errorsCopy)))
	}

	return nil
}

// periodicFlush flushes pending events at regular intervals
func (p *EventProcessor) periodicFlush() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.batchInterval / 2) // Check twice per interval
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.shutdownChan:
			return
		case <-ticker.C:
			if p.shouldFlush() {
				if err := p.Flush(p.ctx); err != nil {
					slog.Error("Periodic flush failed",
						slog.String("error", err.Error()))
				}
			}
		}
	}
}

// periodicMetricsUpdate logs metrics about processing
func (p *EventProcessor) periodicMetricsUpdate() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.metricsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.shutdownChan:
			return
		case <-ticker.C:
			p.mu.Lock()
			buffered := len(p.eventBuffer)
			processed := p.eventsProcessed
			written := p.eventsBatchWritten
			p.mu.Unlock()

			slog.Info("Event processor metrics",
				slog.Int64("events_processed", processed),
				slog.Int64("events_written_to_db", written),
				slog.Int("currently_buffered", buffered))
		}
	}
}

// Shutdown gracefully stops the event processor
func (p *EventProcessor) Shutdown(timeout time.Duration) error {
	slog.Info("Event processor shutdown initiated")

	// Send shutdown signal
	close(p.shutdownChan)

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.cancel()
		slog.Info("Event processor shutdown complete")
		return nil
	case <-time.After(timeout):
		p.cancel()
		return ErrShutdownTimeout
	}
}

// Stats returns current processor statistics
func (p *EventProcessor) Stats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"events_processed":     p.eventsProcessed,
		"events_written_to_db": p.eventsBatchWritten,
		"currently_buffered":   len(p.eventBuffer),
		"pending_errors":       len(p.errorBuffer),
		"batch_size":           p.batchSize,
		"batch_interval_min":   p.batchInterval.Minutes(),
		"last_batch_time":      p.lastBatchTime,
	}
}

// Error definitions
var (
	ErrRedisNotAvailable = fmt.Errorf("redis client not available")
	ErrShutdownTimeout   = fmt.Errorf("event processor shutdown timeout")
)
