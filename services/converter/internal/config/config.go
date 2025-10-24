package config

import (
	"os"
	"strconv"
)

type Config struct {
	// RabbitMQ configuration
	RabbitMQURL      string
	ConversionQueue  string
	NotificationExchange string
	
	// MongoDB configuration
	MongoURL    string
	MongoDB     string
	
	// Redis configuration
	RedisURL string
	
	// Worker configuration
	MaxWorkers     int
	TempDir        string
	
	// FFmpeg configuration
	FFmpegPath     string
	AudioBitrate   string
	AudioSampleRate string
}

func Load() *Config {
	return &Config{
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://admin:dev123@localhost:5672/"),
		ConversionQueue:     getEnv("CONVERSION_QUEUE", "video.conversion"),
		NotificationExchange: getEnv("NOTIFICATION_EXCHANGE", "notifications"),
		
		MongoURL:            getEnv("MONGO_URL", "mongodb://localhost:27017"),
		MongoDB:             getEnv("MONGO_DB", "video_converter"),
		
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
		
		MaxWorkers:          getEnvInt("MAX_WORKERS", 5),
		TempDir:             getEnv("TEMP_DIR", "/tmp/converter"),
		
		FFmpegPath:          getEnv("FFMPEG_PATH", "ffmpeg"),
		AudioBitrate:        getEnv("AUDIO_BITRATE", "192k"),
		AudioSampleRate:     getEnv("AUDIO_SAMPLE_RATE", "44100"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}