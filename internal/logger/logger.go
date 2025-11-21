package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Init initializes the logger based on configuration
func Init(level, format string) {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)

	// Set log level
	switch level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	if format == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}
}

// WithRequestID adds request ID to log entry
func WithRequestID(requestID string) *logrus.Entry {
	return Log.WithField("request_id", requestID)
}

// WithError adds error to log entry
func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
}
