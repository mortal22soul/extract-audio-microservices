package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()

	// JSON format for structured log aggregation (ELK, Loki, CloudWatch, etc.)
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	Log.SetOutput(os.Stdout)

	// Log level from env, defaults to info
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)
}

// WithService returns a log entry pre-tagged with the service name.
func WithService() *logrus.Entry {
	return Log.WithField("service", "gateway")
}

// WithRequestID returns an entry with a request ID for tracing.
func WithRequestID(requestID string) *logrus.Entry {
	return Log.WithField("request_id", requestID)
}
