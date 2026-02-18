package logger_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

func TestLogger(t *testing.T) {

	t.Run("Test Init", func(t *testing.T) {
		logger.Init()
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})

	t.Run("Info level logs", func(t *testing.T) {
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
	})

	t.Run("Debug level logs", func(t *testing.T) {
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
	})

	t.Run("Error level logs", func(t *testing.T) {
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
	})

	t.Run("Warning level logs", func(t *testing.T) {
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

	t.Run("WithRequestID", func(t *testing.T) {
		hook := test.NewGlobal()
		logrus.SetLevel(logrus.InfoLevel)
		defer hook.Reset()

		requestID := "test-request-id-123"
		entry := logger.WithRequestID(requestID)

		assert.NotNil(t, entry)
		assert.Equal(t, requestID, entry.Data["request_id"])
	})

	t.Run("Info level logs with RequestID", func(t *testing.T) {
		t.Run("InfoWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			requestID := "test-request-id-123"
			logger.InfoWithRequestID(requestID, "hello world")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})

		t.Run("InfofWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)
			defer hook.Reset()

			requestID := "test-request-id-456"
			logger.InfofWithRequestID(requestID, "hello %s", "world")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, "hello world", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})
	})

	t.Run("Debug level logs with RequestID", func(t *testing.T) {
		t.Run("DebugWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			requestID := "test-request-id-789"
			logger.DebugWithRequestID(requestID, "debug msg")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug msg", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})

		t.Run("DebugfWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			defer hook.Reset()

			requestID := "test-request-id-101"
			logger.DebugfWithRequestID(requestID, "debug %s", "msg")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.DebugLevel, entry.Level)
			assert.Equal(t, "debug msg", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})
	})

	t.Run("Error level logs with RequestID", func(t *testing.T) {
		t.Run("ErrorWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			requestID := "test-request-id-102"
			logger.ErrorWithRequestID(requestID, "error: not found")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})

		t.Run("ErrorfWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.ErrorLevel)
			defer hook.Reset()

			requestID := "test-request-id-103"
			logger.ErrorfWithRequestID(requestID, "error: %s", "not found")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.Equal(t, "error: not found", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})
	})

	t.Run("Warning level logs with RequestID", func(t *testing.T) {
		t.Run("WarnWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			requestID := "test-request-id-104"
			logger.WarnWithRequestID(requestID, "this is a warning")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
		})

		t.Run("WarnfWithRequestID", func(t *testing.T) {
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.WarnLevel)
			defer hook.Reset()

			requestID := "test-request-id-105"
			logger.WarnfWithRequestID(requestID, "this is a %s", "warning")

			assert.Len(t, hook.Entries, 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Equal(t, "this is a warning", entry.Message)
			assert.Equal(t, requestID, entry.Data["request_id"])
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

		t.Run("FatalWithRequestID", func(t *testing.T) {
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.FatalWithRequestID("request-id-1", "fatal message")
			})
		})

		t.Run("FatalfWithRequestID", func(t *testing.T) {
			assert.PanicsWithValue(t, "fatal-exit", func() {
				logger.FatalfWithRequestID("request-id-2", "fatal %s", "message")
			})
		})
	})
}
