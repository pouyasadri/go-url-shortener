package analytics

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pouyasadri/go-url-shortener/db"
	"github.com/pouyasadri/go-url-shortener/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository provides MongoDB access methods for analytics data
type Repository struct {
	db *mongo.Database
}

// NewRepository creates a new analytics repository
func NewRepository() (*Repository, error) {
	database, err := db.GetDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}
	return &Repository{db: database}, nil
}

// SaveRequestEvent saves a request event to MongoDB
func (r *Repository) SaveRequestEvent(ctx context.Context, event models.RequestEvent) error {
	collection := r.db.Collection("request_events")
	event.CreatedAt = time.Now()

	result, err := collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to save request event: %w", err)
	}

	slog.Debug("Request event saved",
		slog.String("request_id", event.RequestID),
		slog.String("id", fmt.Sprint(result.InsertedID)))
	return nil
}

// SaveRequestEventsBatch saves multiple request events in one operation
func (r *Repository) SaveRequestEventsBatch(ctx context.Context, events []models.RequestEvent) error {
	if len(events) == 0 {
		return nil
	}

	collection := r.db.Collection("request_events")
	now := time.Now()

	// Add timestamp to all events
	docs := make([]interface{}, len(events))
	for i, event := range events {
		if event.CreatedAt.IsZero() {
			event.CreatedAt = now
		}
		docs[i] = event
	}

	opts := options.InsertMany().SetOrdered(false) // Unordered for better performance
	result, err := collection.InsertMany(ctx, docs, opts)
	if err != nil {
		return fmt.Errorf("failed to save request events batch: %w", err)
	}

	slog.Info("Request events batch saved",
		slog.Int("count", len(result.InsertedIDs)))
	return nil
}

// SaveErrorEvent saves an error event to MongoDB
func (r *Repository) SaveErrorEvent(ctx context.Context, event models.ErrorEvent) error {
	collection := r.db.Collection("error_events")
	event.CreatedAt = time.Now()

	result, err := collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to save error event: %w", err)
	}

	slog.Debug("Error event saved",
		slog.String("request_id", event.RequestID),
		slog.String("id", fmt.Sprint(result.InsertedID)))
	return nil
}

// GetRequestCountForHour returns count of requests for a given hour
func (r *Repository) GetRequestCountForHour(ctx context.Context, hour time.Time) (int64, error) {
	collection := r.db.Collection("request_events")

	// Create time range for the hour
	hourStart := hour.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": hourStart,
			"$lt":  hourEnd,
		},
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count requests for hour: %w", err)
	}
	return count, nil
}

// GetLatencyPercentilesForHour calculates p50, p95, p99 latencies for an hour
func (r *Repository) GetLatencyPercentilesForHour(ctx context.Context, hour time.Time) (p50, p95, p99 float64, err error) {
	collection := r.db.Collection("request_events")

	// Create time range for the hour
	hourStart := hour.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": hourStart,
			"$lt":  hourEnd,
		},
	}

	// Aggregate latencies
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "durations", Value: bson.D{
					{Key: "$push", Value: "$duration_ms"},
				}},
			}},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to aggregate latencies: %w", err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 0, 0, 0, nil // No data
	}

	var result struct {
		Durations []float64 `bson:"durations"`
	}
	if err := cursor.Decode(&result); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to decode latencies: %w", err)
	}

	if len(result.Durations) == 0 {
		return 0, 0, 0, nil
	}

	// Sort durations
	sortDurations(result.Durations)

	// Calculate percentiles
	p50 = calculatePercentile(result.Durations, 50)
	p95 = calculatePercentile(result.Durations, 95)
	p99 = calculatePercentile(result.Durations, 99)

	return p50, p95, p99, nil
}

// GetStatusCodeBreakdownForHour returns count of requests by status code for an hour
func (r *Repository) GetStatusCodeBreakdownForHour(ctx context.Context, hour time.Time) (map[string]int64, error) {
	collection := r.db.Collection("request_events")

	// Create time range for the hour
	hourStart := hour.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": hourStart,
			"$lt":  hourEnd,
		},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$status_code"},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate status codes: %w", err)
	}
	defer cursor.Close(ctx)

	breakdown := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    int   `bson:"_id"`
			Count int64 `bson:"count"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode status code: %w", err)
		}
		breakdown[fmt.Sprintf("%d", doc.ID)] = doc.Count
	}

	return breakdown, cursor.Err()
}

// GetErrorCountForHour returns count of errors for an hour
func (r *Repository) GetErrorCountForHour(ctx context.Context, hour time.Time) (int64, error) {
	collection := r.db.Collection("request_events")

	// Create time range for the hour
	hourStart := hour.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": hourStart,
			"$lt":  hourEnd,
		},
		"error_type": bson.M{"$ne": nil}, // Has error
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count errors for hour: %w", err)
	}
	return count, nil
}

// SaveMetricsHourly saves aggregated hourly metrics
func (r *Repository) SaveMetricsHourly(ctx context.Context, metrics models.MetricsHourly) error {
	collection := r.db.Collection("metrics_hourly")
	metrics.CreatedAt = time.Now()
	metrics.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"hour": metrics.Hour}
	update := bson.M{"$set": metrics}

	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save metrics hourly: %w", err)
	}

	slog.Debug("Metrics hourly saved",
		slog.Int64("upserted_id", result.UpsertedCount),
		slog.Int64("modified_count", result.ModifiedCount))
	return nil
}

// GetTotalBytesForHour returns total bytes sent in an hour
func (r *Repository) GetTotalBytesForHour(ctx context.Context, hour time.Time) (int64, error) {
	collection := r.db.Collection("request_events")

	// Create time range for the hour
	hourStart := hour.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": hourStart,
			"$lt":  hourEnd,
		},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total", Value: bson.D{{Key: "$sum", Value: "$bytes_sent"}}},
			}},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate bytes: %w", err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 0, nil
	}

	var result struct {
		Total int64 `bson:"total"`
	}
	if err := cursor.Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode bytes: %w", err)
	}

	return result.Total, nil
}

// DeleteRequestsOlderThan deletes request events older than the given age
func (r *Repository) DeleteRequestsOlderThan(ctx context.Context, age time.Duration) (int64, error) {
	collection := r.db.Collection("request_events")
	cutoff := time.Now().Add(-age)

	filter := bson.M{
		"created_at": bson.M{"$lt": cutoff},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old requests: %w", err)
	}

	slog.Info("Old request events deleted",
		slog.Int64("count", result.DeletedCount),
		slog.Time("before", cutoff))
	return result.DeletedCount, nil
}

// Helper functions

// sortDurations sorts a slice of float64 in ascending order (insertion sort for simplicity)
func sortDurations(durations []float64) {
	for i := 1; i < len(durations); i++ {
		key := durations[i]
		j := i - 1
		for j >= 0 && durations[j] > key {
			durations[j+1] = durations[j]
			j--
		}
		durations[j+1] = key
	}
}

// calculatePercentile returns the value at the given percentile
func calculatePercentile(sorted []float64, percentile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if percentile <= 0 {
		return sorted[0]
	}
	if percentile >= 100 {
		return sorted[len(sorted)-1]
	}

	index := float64(len(sorted)-1) * percentile / 100.0
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[lower]
	}

	// Linear interpolation
	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}
