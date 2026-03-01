package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()

	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	Log.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)
}

// WithService returns a log entry pre-tagged with the service name.
func WithService() *logrus.Entry {
	return Log.WithField("service", "auth")
}

// WithUser returns an entry tagged with a user ID.
func WithUser(userID string) *logrus.Entry {
	return Log.WithField("user_id", userID)
}
