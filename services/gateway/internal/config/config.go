package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	MongoURI          string
	AuthGRPCAddr      string
	AnalyticsGRPCAddr string
	JWTSecret         string
	MaxFileSize       int64
	RateLimit         RateLimitConfig

	// MinIO configuration
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
}

type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

func Load() *Config {
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "104857600"), 10, 64) // 100MB default
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT_RPM", "60"))
	burstSize, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "10"))

	return &Config{
		Port:              getEnv("PORT", "8080"),
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		AuthGRPCAddr:      getEnv("AUTH_GRPC_ADDR", "localhost:50051"),
		AnalyticsGRPCAddr: getEnv("ANALYTICS_GRPC_ADDR", "localhost:50052"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		MaxFileSize:       maxFileSize,
		RateLimit: RateLimitConfig{
			RequestsPerMinute: rateLimit,
			BurstSize:         burstSize,
		},
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "admin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "dev123456"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "videos"),
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}