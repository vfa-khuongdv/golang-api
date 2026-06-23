package logger_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

func TestLogger(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		// Act
		logger.Init(logger.LogConfig{ServiceName: "test"})

		// Assert
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})

	t.Run("Plain logs", func(t *testing.T) {
		t.Run("Info", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			// Act
			logger.Info("hello world")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
		})

		t.Run("Infof", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			// Act
			logger.Infof("hello %s", "world")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
		})

		t.Run("Debug", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			// Act
			logger.Debug("debug msg")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug msg", entry.Message)
		})

		t.Run("Debugf", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			// Act
			logger.Debugf("debug %s", "msg")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug msg", entry.Message)
		})

		t.Run("Error", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			// Act
			logger.Error("error: not found")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
		})

		t.Run("Errorf", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			// Act
			logger.Errorf("error: %s", "not found")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
		})

		t.Run("Warn", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			// Act
			logger.Warn("this is a warning")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
		})

		t.Run("Warnf", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			// Act
			logger.Warnf("this is a %s", "warning")

			// Assert
			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
		})
	})

	t.Run("WithContext logs", func(t *testing.T) {
		t.Run("Infof with requestID", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-123")

			// Act
			logger.WithContext(ctx).Infof("hello %s", "world")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
			assert.Equal(t, "test-req-123", entry.Data["request_id"])
		})

		t.Run("Errorf with requestID", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-456")

			// Act
			logger.WithContext(ctx).Errorf("error: %s", "not found")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
			assert.Equal(t, "test-req-456", entry.Data["request_id"])
		})

		t.Run("Warnf with requestID", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-789")

			// Act
			logger.WithContext(ctx).Warnf("this is a %s", "warning")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
			assert.Equal(t, "test-req-789", entry.Data["request_id"])
		})

		t.Run("WithContext without requestID", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := context.Background()

			// Act
			logger.WithContext(ctx).Infof("no request id")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			_, hasRequestID := entry.Data["request_id"]
			assert.False(t, hasRequestID, "request_id field should not be present when empty")
		})

		t.Run("WithField chaining", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-chain")

			// Act
			logger.WithContext(ctx).WithField("user_id", 42).Infof("user action")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "user action", entry.Message)
			assert.Equal(t, "test-req-chain", entry.Data["request_id"])
			assert.Equal(t, 42, entry.Data["user_id"])
		})

		t.Run("WithFields chaining", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-fields")

			// Act
			logger.WithContext(ctx).WithFields(logrus.Fields{
				"user_id": 99,
				"action":  "login",
			}).Infof("multi fields")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "multi fields", entry.Message)
			assert.Equal(t, "test-req-fields", entry.Data["request_id"])
			assert.Equal(t, 99, entry.Data["user_id"])
			assert.Equal(t, "login", entry.Data["action"])
		})

		t.Run("Logger.Info", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-info")

			// Act
			logger.WithContext(ctx).Info("info message")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "info message", entry.Message)
			assert.Equal(t, "req-info", entry.Data["request_id"])
		})

		t.Run("Logger.Debug", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-debug")

			// Act
			logger.WithContext(ctx).Debug("debug message")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug message", entry.Message)
			assert.Equal(t, "req-debug", entry.Data["request_id"])
		})

		t.Run("Logger.Debugf", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-debugf")

			// Act
			logger.WithContext(ctx).Debugf("debug %s", "formatted")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug formatted", entry.Message)
			assert.Equal(t, "req-debugf", entry.Data["request_id"])
		})

		t.Run("Logger.Error", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-error")

			// Act
			logger.WithContext(ctx).Error("error message")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error message", entry.Message)
			assert.Equal(t, "req-error", entry.Data["request_id"])
		})

		t.Run("Logger.Warn", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-warn")

			// Act
			logger.WithContext(ctx).Warn("warn message")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "warn message", entry.Message)
			assert.Equal(t, "req-warn", entry.Data["request_id"])
		})
	})

	t.Run("Context helpers", func(t *testing.T) {
		t.Run("WithRequestIDContext and RequestIDFromContext", func(t *testing.T) {
			// Arrange
			expectedID := "my-request-id"

			// Act
			ctx := logger.WithRequestIDContext(context.Background(), expectedID)

			// Assert
			assert.Equal(t, expectedID, logger.RequestIDFromContext(ctx))
		})

		t.Run("RequestIDFromContext returns empty for missing key", func(t *testing.T) {
			// Act
			result := logger.RequestIDFromContext(context.Background())

			// Assert
			assert.Equal(t, "", result)
		})
	})

	t.Run("Fatal logs", func(t *testing.T) {
		// Arrange
		originalExitFunc := logrus.StandardLogger().ExitFunc
		logrus.StandardLogger().ExitFunc = func(_ int) { panic("fatal-exit") }
		t.Cleanup(func() {
			logrus.StandardLogger().ExitFunc = originalExitFunc
		})

		t.Run("Fatal", func(t *testing.T) {
			// Assert
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.Fatal("fatal message")
			})
		})

		t.Run("Fatalf", func(t *testing.T) {
			// Assert
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.Fatalf("fatal %s", "message")
			})
		})
	})

	t.Run("Package-level WithField/WithFields", func(t *testing.T) {
		t.Run("WithField", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			// Act
			logger.WithField("request_id", "pkg-req-123").Infof("structured log")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "structured log", entry.Message)
			assert.Equal(t, "pkg-req-123", entry.Data["request_id"])
		})

		t.Run("WithFields", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			// Act
			logger.WithFields(logrus.Fields{
				"request_id": "pkg-req-456",
				"component":  "middleware",
			}).Infof("multi-field log")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "multi-field log", entry.Message)
			assert.Equal(t, "pkg-req-456", entry.Data["request_id"])
			assert.Equal(t, "middleware", entry.Data["component"])
		})

		t.Run("WithField Errorf", func(t *testing.T) {
			// Arrange
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			// Act
			logger.WithField("request_id", "pkg-req-789").Errorf("error: %s", "something failed")

			// Assert
			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: something failed", entry.Message)
			assert.Equal(t, "pkg-req-789", entry.Data["request_id"])
		})
	})

	t.Run("WithEvent", func(t *testing.T) {
		// Arrange
		hook := test.NewGlobal()
		logrus.SetLevel(logrus.InfoLevel)
		defer hook.Reset()

		ctx := logger.WithRequestIDContext(context.Background(), "req-event")

		// Act
		logger.WithEvent(ctx, logger.EventLoginAttempt).Infof("test login")

		// Assert
		require.Len(t, hook.Entries, 1)
		entry := hook.LastEntry()
		assert.Equal(t, "test login", entry.Message)
		assert.Equal(t, logger.EventLoginAttempt, entry.Data["event"])
		assert.Equal(t, "req-event", entry.Data["request_id"])
	})

	t.Run("WithEventAndFields", func(t *testing.T) {
		// Arrange
		hook := test.NewGlobal()
		logrus.SetLevel(logrus.InfoLevel)
		defer hook.Reset()

		ctx := logger.WithRequestIDContext(context.Background(), "req-event-fields")

		// Act
		logger.WithEventAndFields(ctx, logger.EventLoginSuccess, logrus.Fields{
			"user_id": 42,
		}).Infof("test success")

		// Assert
		require.Len(t, hook.Entries, 1)
		entry := hook.LastEntry()
		assert.Equal(t, "test success", entry.Message)
		assert.Equal(t, logger.EventLoginSuccess, entry.Data["event"])
		assert.Equal(t, 42, entry.Data["user_id"])
		assert.Equal(t, "req-event-fields", entry.Data["request_id"])
	})

	t.Run("Init with full config", func(t *testing.T) {
		// Act
		logger.Init(logger.LogConfig{
			ServiceName: "test-service",
			Stage:       "staging",
			Version:     "2.0.0",
		})

		// Assert
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})
}
