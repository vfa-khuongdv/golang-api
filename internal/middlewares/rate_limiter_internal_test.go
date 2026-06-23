package middlewares

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanup(t *testing.T) {
	t.Run("removes expired entries, keeps valid ones", func(t *testing.T) {
		// Arrange
		rl := &rateLimiter{
			requests:        make(map[string][]time.Time),
			limit:           5,
			window:          50 * time.Millisecond,
			cleanupInterval: time.Minute,
			stopCh:          make(chan struct{}),
		}
		now := time.Now()
		rl.requests["key1"] = []time.Time{now.Add(-100 * time.Millisecond), now.Add(-10 * time.Millisecond)}

		// Act
		rl.cleanup()

		// Assert
		rl.mu.Lock()
		entries := rl.requests["key1"]
		rl.mu.Unlock()
		assert.Len(t, entries, 1)
		assert.True(t, entries[0].After(now.Add(-50*time.Millisecond)))
	})

	t.Run("deletes key when no valid entries remain", func(t *testing.T) {
		// Arrange
		rl := &rateLimiter{
			requests:        make(map[string][]time.Time),
			limit:           5,
			window:          50 * time.Millisecond,
			cleanupInterval: time.Minute,
			stopCh:          make(chan struct{}),
		}
		rl.requests["expired"] = []time.Time{time.Now().Add(-time.Hour)}

		// Act
		rl.cleanup()

		// Assert
		rl.mu.Lock()
		_, exists := rl.requests["expired"]
		rl.mu.Unlock()
		assert.False(t, exists)
	})
}

func TestCleanupLoopStop(t *testing.T) {
	// Arrange
	rl := &rateLimiter{
		requests:        make(map[string][]time.Time),
		limit:           5,
		window:          time.Second,
		cleanupInterval: time.Minute,
		stopCh:          make(chan struct{}),
	}
	var wg sync.WaitGroup
	wg.Add(1)

	// Act
	go func() {
		defer wg.Done()
		rl.cleanupLoop()
	}()

	time.Sleep(time.Millisecond)
	close(rl.stopCh)
	wg.Wait()
}

func TestCleanupLoopTicker(t *testing.T) {
	// Arrange
	rl := &rateLimiter{
		requests:        make(map[string][]time.Time),
		limit:           5,
		window:          10 * time.Millisecond,
		cleanupInterval: 10 * time.Millisecond,
		stopCh:          make(chan struct{}),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rl.cleanupLoop()
	}()

	rl.mu.Lock()
	rl.requests["ticker_test"] = []time.Time{time.Now().Add(-time.Hour)}
	rl.mu.Unlock()

	// Act
	time.Sleep(30 * time.Millisecond)

	// Assert
	rl.mu.Lock()
	_, exists := rl.requests["ticker_test"]
	rl.mu.Unlock()
	assert.False(t, exists, "expired entry should have been cleaned up by ticker")

	close(rl.stopCh)
	wg.Wait()
}
