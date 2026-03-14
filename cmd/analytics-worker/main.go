package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pouyasadri/go-url-shortener/analytics"
	"github.com/pouyasadri/go-url-shortener/config"
	"github.com/pouyasadri/go-url-shortener/store"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Analytics Worker Service Starting")

	// Initialize store (Redis)
	store.InitializeStore()

	// Load analytics config
	analyticsCfg := config.LoadAnalyticsConfig()

	if !analyticsCfg.Enabled {
		slog.Warn("Analytics is disabled, worker exiting")
		return
	}

	// Create analytics repository
	repo, err := analytics.NewRepository()
	if err != nil {
		log.Fatalf("Failed to create analytics repository: %v", err)
	}

	slog.Info("Analytics repository initialized")

	// Create and start event processor
	eventProcessor := NewEventProcessor(repo, 1000, analyticsCfg.BatchWriteIntervalMinutes)
	if err := eventProcessor.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start event processor: %v", err)
	}

	// Create and start metrics aggregator
	metricsAggregator := NewMetricsAggregator(repo, analyticsCfg.AggregationIntervalMinutes)
	if err := metricsAggregator.Start(); err != nil {
		log.Fatalf("Failed to start metrics aggregator: %v", err)
	}

	// Create and start cleanup job
	cleanupJob := NewCleanupJob(repo, analyticsCfg.MetricsRetentionDays)
	if err := cleanupJob.Start(); err != nil {
		log.Fatalf("Failed to start cleanup job: %v", err)
	}

	slog.Info("All worker components started successfully")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Log health status periodically
	go func() {
		for range time.Tick(30 * time.Second) {
			stats := eventProcessor.Stats()
			slog.Info("Worker health check",
				slog.Int64("events_processed", stats["events_processed"].(int64)),
				slog.Int64("events_written_to_db", stats["events_written_to_db"].(int64)),
				slog.Int("currently_buffered", stats["currently_buffered"].(int)))
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	slog.Info("Shutdown signal received",
		slog.String("signal", sig.String()))

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = ctx // Unused for now, but reserved for future use

	slog.Info("Shutting down components")

	// Stop cleanup job
	cleanupJob.Stop()

	// Stop metrics aggregator
	metricsAggregator.Stop()

	// Shutdown event processor
	if err := eventProcessor.Shutdown(15 * time.Second); err != nil {
		slog.Error("Error during event processor shutdown",
			slog.String("error", err.Error()))
	}

	slog.Info("Analytics Worker Service Stopped")
}
