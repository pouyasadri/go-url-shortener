package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
	once   sync.Once
)

// GetMongoClient returns the MongoDB client singleton
func GetMongoClient() (*mongo.Client, error) {
	var err error
	once.Do(func() {
		client, err = createMongoClient()
	})
	return client, err
}

// createMongoClient creates a new MongoDB client with connection pooling
func createMongoClient() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := getMongoURI()
	opts := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	slog.Info("Connected to MongoDB", slog.String("uri", mongoURI))
	return client, nil
}

// GetDatabase returns the MongoDB database
func GetDatabase() (*mongo.Database, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}
	return client.Database(getDatabaseName()), nil
}

// Disconnect closes the MongoDB connection
func Disconnect(ctx context.Context) error {
	if client == nil {
		return nil
	}
	return client.Disconnect(ctx)
}

// HealthCheck verifies the MongoDB connection is healthy
func HealthCheck(ctx context.Context) error {
	client, err := GetMongoClient()
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return client.Ping(timeoutCtx, nil)
}

// Helper functions to get configuration from environment
func getMongoURI() string {
	uri := getEnv("MONGODB_URI", "mongodb://mongo:27017")
	return uri
}

func getDatabaseName() string {
	return getEnv("MONGODB_DB", "url_shortener")
}

func getEnv(key, defaultVal string) string {
	// This will be replaced with proper config package integration
	// For now, using simple getenv
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	return val
}
