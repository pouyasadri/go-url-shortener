package config

// AnalyticsConfig holds all analytics-related configuration
type AnalyticsConfig struct {
	// Enabled determines if analytics collection is active
	Enabled bool

	// MetricsRetentionDays is how long to keep raw request events (30 days default)
	MetricsRetentionDays int

	// BatchWriteIntervalMinutes is how long to buffer events before writing to MongoDB
	BatchWriteIntervalMinutes int

	// AggregationIntervalMinutes is how often to aggregate metrics
	AggregationIntervalMinutes int

	// MongoDBURI is the connection string for MongoDB
	MongoDBURI string

	// MongoDBName is the database name
	MongoDBName string
}

// LoadAnalyticsConfig loads analytics configuration from environment variables
func LoadAnalyticsConfig() *AnalyticsConfig {
	return &AnalyticsConfig{
		Enabled:                    getEnvBool("ANALYTICS_ENABLED", true),
		MetricsRetentionDays:       getEnvInt("METRICS_RETENTION_DAYS", 30),
		BatchWriteIntervalMinutes:  getEnvInt("BATCH_WRITE_INTERVAL_MINUTES", 5),
		AggregationIntervalMinutes: getEnvInt("AGGREGATION_INTERVAL_MINUTES", 60),
		MongoDBURI:                 getEnv("MONGODB_URI", "mongodb://mongo:27017"),
		MongoDBName:                getEnv("MONGODB_DB", "url_shortener"),
	}
}
