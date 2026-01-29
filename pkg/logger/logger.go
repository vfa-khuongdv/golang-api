package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func Init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

}

// WithRequestID creates a new log entry with requestID field
// Use this for logging within a request context to trace logs
func WithRequestID(requestID string) *log.Entry {
	return log.WithField("request_id", requestID)
}

// InfoWithRequestID logs information level messages with requestID
func InfoWithRequestID(requestID string, args ...interface{}) {
	WithRequestID(requestID).Info(args...)
}

// InfofWithRequestID logs formatted information level messages with requestID
func InfofWithRequestID(requestID string, format string, args ...interface{}) {
	WithRequestID(requestID).Infof(format, args...)
}

// DebugWithRequestID logs debug level messages with requestID
func DebugWithRequestID(requestID string, args ...interface{}) {
	WithRequestID(requestID).Debug(args...)
}

// DebugfWithRequestID logs formatted debug level messages with requestID
func DebugfWithRequestID(requestID string, format string, args ...interface{}) {
	WithRequestID(requestID).Debugf(format, args...)
}

// ErrorWithRequestID logs error level messages with requestID
func ErrorWithRequestID(requestID string, args ...interface{}) {
	WithRequestID(requestID).Error(args...)
}

// ErrorfWithRequestID logs formatted error level messages with requestID
func ErrorfWithRequestID(requestID string, format string, args ...interface{}) {
	WithRequestID(requestID).Errorf(format, args...)
}

// WarnWithRequestID logs warning level messages with requestID
func WarnWithRequestID(requestID string, args ...interface{}) {
	WithRequestID(requestID).Warn(args...)
}

// WarnfWithRequestID logs formatted warning level messages with requestID
func WarnfWithRequestID(requestID string, format string, args ...interface{}) {
	WithRequestID(requestID).Warnf(format, args...)
}

// FatalWithRequestID logs fatal level messages with requestID and then exits
func FatalWithRequestID(requestID string, args ...interface{}) {
	WithRequestID(requestID).Fatal(args...)
}

// FatalfWithRequestID logs formatted fatal level messages with requestID and then exits
func FatalfWithRequestID(requestID string, format string, args ...interface{}) {
	WithRequestID(requestID).Fatalf(format, args...)
}

// Info logs information level messages
// Takes variadic arguments and logs them at INFO level
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof logs formatted information level messages
// Takes a format string and variadic arguments to format the log message
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Debug logs debug level messages
// Takes variadic arguments and logs them at DEBUG level
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf logs formatted debug level messages
// Takes a format string and variadic arguments to format the log message
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Error logs error level messages
// Takes variadic arguments and logs them at ERROR level
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf logs formatted error level messages
// Takes a format string and variadic arguments to format the log message
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal logs fatal level messages and then exits with status code 1
// Takes variadic arguments, logs them at FATAL level and terminates program
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf logs formatted fatal level messages and then exits
// Takes a format string and variadic arguments, formats and logs at FATAL level before terminating
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Warn logs warning level messages
// Takes variadic arguments and logs them at WARN level
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf logs formatted warning level messages
// Takes a format string and variadic arguments to format the log message
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}
