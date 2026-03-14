package db

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitializeIndexes creates all necessary MongoDB indexes
func InitializeIndexes(ctx context.Context) error {
	db, err := GetDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Create indexes for each collection
	if err := createRequestEventIndexes(ctx, db); err != nil {
		return err
	}

	if err := createErrorEventIndexes(ctx, db); err != nil {
		return err
	}

	if err := createMetricsHourlyIndexes(ctx, db); err != nil {
		return err
	}

	if err := createUserAnalyticsIndexes(ctx, db); err != nil {
		return err
	}

	if err := createAPIKeyAnalyticsIndexes(ctx, db); err != nil {
		return err
	}

	if err := createURLAnalyticsIndexes(ctx, db); err != nil {
		return err
	}

	slog.Info("All MongoDB indexes initialized")
	return nil
}

// createRequestEventIndexes creates indexes for request_events collection
func createRequestEventIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("request_events")
	ttlSeconds := int32(30 * 24 * 60 * 60) // 30 days in seconds
	indexModel := []mongo.IndexModel{
		// TTL index: auto-delete documents older than 30 days
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(ttlSeconds),
		},
		// Timestamp index for querying by time range
		{
			Keys: bson.D{{Key: "timestamp", Value: -1}},
		},
		// Request ID index for quick lookups
		{
			Keys:    bson.D{{Key: "request_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// API key index for filtering by api_key_id
		{
			Keys: bson.D{{Key: "api_key_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		// User ID index for user analytics
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		// Status code index for breakdown analysis
		{
			Keys: bson.D{{Key: "status_code", Value: 1}},
		},
		// Path index for path-based aggregations
		{
			Keys: bson.D{{Key: "path", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create request_events indexes: %w", err)
	}

	slog.Info("Created indexes for request_events collection")
	return nil
}

// createErrorEventIndexes creates indexes for error_events collection
func createErrorEventIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("error_events")
	ttlSeconds := int32(30 * 24 * 60 * 60) // 30 days in seconds
	indexModel := []mongo.IndexModel{
		// TTL index: auto-delete documents older than 30 days
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(ttlSeconds),
		},
		// Timestamp index for time-range queries
		{
			Keys: bson.D{{Key: "timestamp", Value: -1}},
		},
		// Request ID index for correlation
		{
			Keys:    bson.D{{Key: "request_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// Error type index for breakdown
		{
			Keys: bson.D{{Key: "error_type", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		// API key index
		{
			Keys: bson.D{{Key: "api_key_id", Value: 1}},
		},
		// User ID index
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create error_events indexes: %w", err)
	}

	slog.Info("Created indexes for error_events collection")
	return nil
}

// createMetricsHourlyIndexes creates indexes for metrics_hourly collection
func createMetricsHourlyIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("metrics_hourly")
	indexModel := []mongo.IndexModel{
		// Hour index for querying by time period (unique to prevent duplicates)
		{
			Keys:    bson.D{{Key: "hour", Value: -1}},
			Options: options.Index().SetUnique(true),
		},
		// Created at index for cleanup
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create metrics_hourly indexes: %w", err)
	}

	slog.Info("Created indexes for metrics_hourly collection")
	return nil
}

// createUserAnalyticsIndexes creates indexes for user_analytics collection
func createUserAnalyticsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("user_analytics")
	indexModel := []mongo.IndexModel{
		// User ID index (unique)
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// Last active index for sorting
		{
			Keys: bson.D{{Key: "last_active", Value: -1}},
		},
		// API calls index for sorting
		{
			Keys: bson.D{{Key: "api_calls_total", Value: -1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create user_analytics indexes: %w", err)
	}

	slog.Info("Created indexes for user_analytics collection")
	return nil
}

// createAPIKeyAnalyticsIndexes creates indexes for api_key_analytics collection
func createAPIKeyAnalyticsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("api_key_analytics")
	indexModel := []mongo.IndexModel{
		// API key ID index (unique)
		{
			Keys:    bson.D{{Key: "api_key_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// Last used index for sorting
		{
			Keys: bson.D{{Key: "last_used", Value: -1}},
		},
		// Usage count index for sorting
		{
			Keys: bson.D{{Key: "usage_count", Value: -1}},
		},
		// Error rate index for health monitoring
		{
			Keys: bson.D{{Key: "error_rate", Value: -1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create api_key_analytics indexes: %w", err)
	}

	slog.Info("Created indexes for api_key_analytics collection")
	return nil
}

// createURLAnalyticsIndexes creates indexes for url_analytics collection
func createURLAnalyticsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("url_analytics")
	indexModel := []mongo.IndexModel{
		// Short URL index (unique)
		{
			Keys:    bson.D{{Key: "short_url", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// Created by index for user's URLs
		{
			Keys: bson.D{{Key: "created_by", Value: 1}, {Key: "created_at", Value: -1}},
		},
		// Redirect count index for popular URLs
		{
			Keys: bson.D{{Key: "redirect_count", Value: -1}},
		},
		// Updated at index for recent activity
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create url_analytics indexes: %w", err)
	}

	slog.Info("Created indexes for url_analytics collection")
	return nil
}
