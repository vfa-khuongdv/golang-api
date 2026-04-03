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
		logger.Init()
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})

	t.Run("Plain logs", func(t *testing.T) {
		t.Run("Info", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			logger.Info("hello world")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
		})

		t.Run("Infof", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			logger.Infof("hello %s", "world")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
		})

		t.Run("Debug", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			logger.Debug("debug msg")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug msg", entry.Message)
		})

		t.Run("Debugf", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			logger.Debugf("debug %s", "msg")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug msg", entry.Message)
		})

		t.Run("Error", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			logger.Error("error: not found")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
		})

		t.Run("Errorf", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			logger.Errorf("error: %s", "not found")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
		})

		t.Run("Warn", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			logger.Warn("this is a warning")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
		})

		t.Run("Warnf", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			logger.Warnf("this is a %s", "warning")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
		})
	})

	t.Run("WithContext logs", func(t *testing.T) {
		t.Run("Infof with requestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-123")
			logger.WithContext(ctx).Infof("hello %s", "world")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
			assert.Equal(t, "test-req-123", entry.Data["request_id"])
		})

		t.Run("Errorf with requestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-456")
			logger.WithContext(ctx).Errorf("error: %s", "not found")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
			assert.Equal(t, "test-req-456", entry.Data["request_id"])
		})

		t.Run("Warnf with requestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-789")
			logger.WithContext(ctx).Warnf("this is a %s", "warning")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
			assert.Equal(t, "test-req-789", entry.Data["request_id"])
		})

		t.Run("WithContext without requestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := context.Background()
			logger.WithContext(ctx).Infof("no request id")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "", entry.Data["request_id"])
		})

		t.Run("WithField chaining", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-chain")
			logger.WithContext(ctx).WithField("user_id", 42).Infof("user action")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "user action", entry.Message)
			assert.Equal(t, "test-req-chain", entry.Data["request_id"])
			assert.Equal(t, 42, entry.Data["user_id"])
		})

		t.Run("WithFields chaining", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "test-req-fields")
			logger.WithContext(ctx).WithFields(logrus.Fields{
				"user_id": 99,
				"action":  "login",
			}).Infof("multi fields")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, "multi fields", entry.Message)
			assert.Equal(t, "test-req-fields", entry.Data["request_id"])
			assert.Equal(t, 99, entry.Data["user_id"])
			assert.Equal(t, "login", entry.Data["action"])
		})

		t.Run("Logger.Info", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-info")
			logger.WithContext(ctx).Info("info message")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "info message", entry.Message)
			assert.Equal(t, "req-info", entry.Data["request_id"])
		})

		t.Run("Logger.Debug", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-debug")
			logger.WithContext(ctx).Debug("debug message")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug message", entry.Message)
			assert.Equal(t, "req-debug", entry.Data["request_id"])
		})

		t.Run("Logger.Debugf", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-debugf")
			logger.WithContext(ctx).Debugf("debug %s", "formatted")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug formatted", entry.Message)
			assert.Equal(t, "req-debugf", entry.Data["request_id"])
		})

		t.Run("Logger.Error", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-error")
			logger.WithContext(ctx).Error("error message")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error message", entry.Message)
			assert.Equal(t, "req-error", entry.Data["request_id"])
		})

		t.Run("Logger.Warn", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-warn")
			logger.WithContext(ctx).Warn("warn message")

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "warn message", entry.Message)
			assert.Equal(t, "req-warn", entry.Data["request_id"])
		})
	})

	t.Run("Context helpers", func(t *testing.T) {
		t.Run("WithRequestIDContext and RequestIDFromContext", func(t *testing.T) {
			ctx := logger.WithRequestIDContext(context.Background(), "my-request-id")
			assert.Equal(t, "my-request-id", logger.RequestIDFromContext(ctx))
		})

		t.Run("RequestIDFromContext returns empty for missing key", func(t *testing.T) {
			assert.Equal(t, "", logger.RequestIDFromContext(context.Background()))
		})
	})

	t.Run("Fatal logs", func(t *testing.T) {
		originalExitFunc := logrus.StandardLogger().ExitFunc
		logrus.StandardLogger().ExitFunc = func(_ int) { panic("fatal-exit") }
		t.Cleanup(func() {
			logrus.StandardLogger().ExitFunc = originalExitFunc
		})

		t.Run("Fatal", func(t *testing.T) {
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.Fatal("fatal message")
			})
		})

		t.Run("Fatalf", func(t *testing.T) {
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.Fatalf("fatal %s", "message")
			})
		})

		t.Run("Logger.Fatal", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.FatalLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-fatal")
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.WithContext(ctx).Fatal("fatal from context")
			})

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.FatalLevel, entry.Level)
			assert.Equal(t, "fatal from context", entry.Message)
			assert.Equal(t, "req-fatal", entry.Data["request_id"])
		})

		t.Run("Logger.Fatalf", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.FatalLevel)
			defer hook.Reset()

			ctx := logger.WithRequestIDContext(context.Background(), "req-fatalf")
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.WithContext(ctx).Fatalf("fatal %s", "formatted")
			})

			require.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.FatalLevel, entry.Level)
			assert.Equal(t, "fatal formatted", entry.Message)
			assert.Equal(t, "req-fatalf", entry.Data["request_id"])
		})
	})
}
