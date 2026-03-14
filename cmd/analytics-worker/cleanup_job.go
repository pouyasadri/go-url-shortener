package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pouyasadri/go-url-shortener/analytics"
	"github.com/robfig/cron/v3"
)

// CleanupJob performs periodic cleanup of old data
type CleanupJob struct {
	repo             *analytics.Repository
	cron             *cron.Cron
	retentionDays    int
	cleanupHourOfDay int // 0-23, hour of day to run cleanup
}

// NewCleanupJob creates a new cleanup job
func NewCleanupJob(repo *analytics.Repository, retentionDays int) *CleanupJob {
	return &CleanupJob{
		repo:             repo,
		cron:             cron.New(),
		retentionDays:    retentionDays,
		cleanupHourOfDay: 2, // Run cleanup at 2 AM
	}
}

// Start begins the daily cleanup job
func (cj *CleanupJob) Start() error {
	slog.Info("Cleanup job starting",
		slog.Int("retention_days", cj.retentionDays),
		slog.Int("cleanup_hour", cj.cleanupHourOfDay))

	// Schedule daily at 2 AM (format: minute hour day month dayOfWeek)
	cronExpr := fmt.Sprintf("0 %d * * *", cj.cleanupHourOfDay)
	_, err := cj.cron.AddFunc(cronExpr, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := cj.runCleanup(ctx); err != nil {
			slog.Error("Cleanup job failed",
				slog.String("error", err.Error()))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to schedule cleanup job: %w", err)
	}

	cj.cron.Start()
	slog.Info("Cleanup job scheduled")
	return nil
}

// runCleanup performs the actual cleanup
func (cj *CleanupJob) runCleanup(ctx context.Context) error {
	slog.Info("Running cleanup job",
		slog.Int("retention_days", cj.retentionDays))

	age := time.Duration(cj.retentionDays) * 24 * time.Hour

	// Delete old request events
	deletedRequests, err := cj.repo.DeleteRequestsOlderThan(ctx, age)
	if err != nil {
		return fmt.Errorf("failed to delete old requests: %w", err)
	}

	slog.Info("Cleanup job completed",
		slog.Int64("requests_deleted", deletedRequests),
		slog.Int("retention_days", cj.retentionDays))

	return nil
}

// Stop stops the cleanup job
func (cj *CleanupJob) Stop() {
	cj.cron.Stop()
	slog.Info("Cleanup job stopped")
}

// TriggerNow forces an immediate cleanup (for testing)
func (cj *CleanupJob) TriggerNow(ctx context.Context) error {
	return cj.runCleanup(ctx)
}
