package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitMQ RabbitMQConfig
	SMTP     SMTPConfig
	Service  ServiceConfig
}

type RabbitMQConfig struct {
	URL       string
	QueueName string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type ServiceConfig struct {
	Port string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	
	return &Config{
		RabbitMQ: RabbitMQConfig{
			URL:       getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			QueueName: getEnv("NOTIFICATION_QUEUE", "notifications"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     port,
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@videoconverter.com"),
		},
		Service: ServiceConfig{
			Port: getEnv("PORT", "8080"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}