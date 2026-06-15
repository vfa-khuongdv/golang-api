package logger

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// Event type constants for structured logging.
// Used as the "event" field in every log entry for easy filtering in ELK/Kibana.
const (
	EventLoginAttempt            = "login_attempt"
	EventLoginSuccess            = "login_success"
	EventLoginFailed             = "login_failed"
	EventTokenRefresh            = "token_refresh"
	EventTokenRefreshSuccess     = "token_refresh_success"
	EventTokenRefreshFailed      = "token_refresh_failed"
	EventPasswordResetRequest    = "password_reset_request"
	EventPasswordReset           = "password_reset"
	EventPasswordChange          = "password_change"
	EventPasswordChangeFailed    = "password_change_failed"
	EventProfileGet              = "profile_get"
	EventProfileUpdate           = "profile_update"
	EventProfileUpdateFailed     = "profile_update_failed"
)

// Logger wraps a logrus entry for structured logging
type Logger struct {
	entry *log.Entry
}

// WithContext returns a Logger with requestID extracted from context.
// If context has no requestID, logs without request_id field.
func WithContext(ctx context.Context) Logger {
	requestID := ""
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		requestID = id
	}
	if requestID != "" {
		return Logger{entry: log.WithField("request_id", requestID)}
	}
	return Logger{entry: log.NewEntry(log.StandardLogger())}
}

// WithField returns a new Logger with an additional field
func (l Logger) WithField(key string, value interface{}) Logger {
	return Logger{entry: l.entry.WithField(key, value)}
}

// WithFields returns a new Logger with additional fields
func (l Logger) WithFields(fields log.Fields) Logger {
	return Logger{entry: l.entry.WithFields(fields)}
}

func (l Logger) Info(args ...interface{})                  { l.entry.Info(args...) }
func (l Logger) Infof(format string, args ...interface{})  { l.entry.Infof(format, args...) }
func (l Logger) Debug(args ...interface{})                 { l.entry.Debug(args...) }
func (l Logger) Debugf(format string, args ...interface{}) { l.entry.Debugf(format, args...) }
func (l Logger) Error(args ...interface{})                 { l.entry.Error(args...) }
func (l Logger) Errorf(format string, args ...interface{}) { l.entry.Errorf(format, args...) }
func (l Logger) Warn(args ...interface{})                  { l.entry.Warn(args...) }
func (l Logger) Warnf(format string, args ...interface{})  { l.entry.Warnf(format, args...) }

// RequestIDContext helpers

func WithRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

type LogConfig struct {
	ServiceName string
	Stage       string
	Version     string
}

// Init configures the global logger with standard fields.
// serviceName, stage, and version are added to every log entry.
func Init(cfg LogConfig) {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg: "message",
		},
	})
	log.SetOutput(os.Stdout)

	defaultFields := log.Fields{
		"service": cfg.ServiceName,
		"version": cfg.Version,
	}
	if cfg.Stage != "" {
		defaultFields["env"] = cfg.Stage
	}

	if logEntry := log.WithFields(defaultFields); cfg.Version != "" {
		logEntry.Info("starting")
	}
}

// Plain log functions (no context, no requestID) for startup, seeders, etc.

func Info(args ...interface{})                  { log.Info(args...) }
func Infof(format string, args ...interface{})  { log.Infof(format, args...) }
func Debug(args ...interface{})                 { log.Debug(args...) }
func Debugf(format string, args ...interface{}) { log.Debugf(format, args...) }
func Error(args ...interface{})                 { log.Error(args...) }
func Errorf(format string, args ...interface{}) { log.Errorf(format, args...) }
func Warn(args ...interface{})                  { log.Warn(args...) }
func Warnf(format string, args ...interface{})  { log.Warnf(format, args...) }
func Fatal(args ...interface{})                 { log.Fatal(args...) }
func Fatalf(format string, args ...interface{}) { log.Fatalf(format, args...) }

// WithField returns a Logger with a single field for non-context structured logging
func WithField(key string, value interface{}) Logger {
	return Logger{entry: log.WithField(key, value)}
}

// WithFields returns a Logger with multiple fields for non-context structured logging
func WithFields(fields log.Fields) Logger {
	return Logger{entry: log.WithFields(fields)}
}

// WithEvent returns a Logger with an "event" field set.
// Use for consistent event tracking across ELK/Kibana.
func WithEvent(ctx context.Context, event string) Logger {
	return WithContext(ctx).WithField("event", event)
}

// WithEventAndFields returns a Logger with event and extra fields.
func WithEventAndFields(ctx context.Context, event string, fields log.Fields) Logger {
	f := log.Fields{"event": event}
	for k, v := range fields {
		f[k] = v
	}
	return Logger{entry: WithContext(ctx).entry.WithFields(f)}
}
