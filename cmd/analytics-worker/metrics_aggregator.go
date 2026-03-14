package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pouyasadri/go-url-shortener/analytics"
	"github.com/pouyasadri/go-url-shortener/models"
	"github.com/robfig/cron/v3"
)

// MetricsAggregator calculates and stores hourly metrics
type MetricsAggregator struct {
	repo     *analytics.Repository
	cron     *cron.Cron
	interval time.Duration
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator(repo *analytics.Repository, intervalMinutes int) *MetricsAggregator {
	return &MetricsAggregator{
		repo:     repo,
		cron:     cron.New(),
		interval: time.Duration(intervalMinutes) * time.Minute,
	}
}

// Start begins the hourly metrics aggregation job
func (ma *MetricsAggregator) Start() error {
	slog.Info("Metrics aggregator starting",
		slog.Duration("interval", ma.interval))

	// Schedule every hour
	_, err := ma.cron.AddFunc("0 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := ma.aggregateLastHour(ctx); err != nil {
			slog.Error("Failed to aggregate metrics",
				slog.String("error", err.Error()))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to schedule aggregation job: %w", err)
	}

	ma.cron.Start()
	slog.Info("Metrics aggregator scheduled")
	return nil
}

// aggregateLastHour calculates metrics for the previous hour
func (ma *MetricsAggregator) aggregateLastHour(ctx context.Context) error {
	// Get the last completed hour
	now := time.Now()
	hour := now.Truncate(time.Hour).Add(-time.Hour)

	slog.Info("Aggregating metrics for hour",
		slog.Time("hour", hour))

	// Get counts
	requestCount, err := ma.repo.GetRequestCountForHour(ctx, hour)
	if err != nil {
		return fmt.Errorf("failed to get request count: %w", err)
	}

	if requestCount == 0 {
		slog.Debug("No requests for hour, skipping aggregation",
			slog.Time("hour", hour))
		return nil
	}

	// Get error count
	errorCount, err := ma.repo.GetErrorCountForHour(ctx, hour)
	if err != nil {
		slog.Warn("Failed to get error count",
			slog.String("error", err.Error()))
		errorCount = 0
	}

	// Get latency percentiles
	p50, p95, p99, err := ma.repo.GetLatencyPercentilesForHour(ctx, hour)
	if err != nil {
		slog.Warn("Failed to get latency percentiles",
			slog.String("error", err.Error()))
	}

	// Get status code breakdown
	statusCodeBreakdown, err := ma.repo.GetStatusCodeBreakdownForHour(ctx, hour)
	if err != nil {
		slog.Warn("Failed to get status code breakdown",
			slog.String("error", err.Error()))
		statusCodeBreakdown = make(map[string]int64)
	}

	// Get total bytes
	totalBytes, err := ma.repo.GetTotalBytesForHour(ctx, hour)
	if err != nil {
		slog.Warn("Failed to get total bytes",
			slog.String("error", err.Error()))
		totalBytes = 0
	}

	// Calculate average latency
	avgLatency := 0.0
	if requestCount > 0 {
		avgLatency = (p50 + p95 + p99) / 3.0
	}

	// Create and save metrics
	metrics := models.MetricsHourly{
		Hour:               hour,
		RequestCount:       requestCount,
		ErrorCount:         errorCount,
		AvgLatencyMs:       avgLatency,
		P50LatencyMs:       p50,
		P95LatencyMs:       p95,
		P99LatencyMs:       p99,
		TotalBytesSent:     totalBytes,
		StatusCodeBreak:    statusCodeBreakdown,
		PathBreakdown:      make(map[string]int64), // TODO: implement path breakdown
		ErrorTypeBreakdown: make(map[string]int64), // TODO: implement error type breakdown
	}

	if err := ma.repo.SaveMetricsHourly(ctx, metrics); err != nil {
		return fmt.Errorf("failed to save metrics: %w", err)
	}

	slog.Info("Metrics aggregated for hour",
		slog.Time("hour", hour),
		slog.Int64("request_count", requestCount),
		slog.Int64("error_count", errorCount),
		slog.Float64("p50_latency_ms", p50),
		slog.Float64("p95_latency_ms", p95),
		slog.Float64("p99_latency_ms", p99))

	return nil
}

// Stop stops the metrics aggregation job
func (ma *MetricsAggregator) Stop() {
	ma.cron.Stop()
	slog.Info("Metrics aggregator stopped")
}

// TriggerNow forces an immediate aggregation (for testing)
func (ma *MetricsAggregator) TriggerNow(ctx context.Context) error {
	return ma.aggregateLastHour(ctx)
}
