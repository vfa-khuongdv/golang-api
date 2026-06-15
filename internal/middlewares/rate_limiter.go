package middlewares

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type rateLimiter struct {
	requests        map[string][]time.Time
	mu              sync.Mutex
	limit           int
	window          time.Duration
	cleanupInterval time.Duration
	stopCh          chan struct{}
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests:        make(map[string][]time.Time),
		limit:           limit,
		window:          window,
		cleanupInterval: window * 2,
		stopCh:          make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	windowStart := time.Now().Add(-rl.window)
	for key, timestamps := range rl.requests {
		valid := timestamps[:0]
		for _, t := range timestamps {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

func (rl *rateLimiter) allow(key string) (bool, int, int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	valid := rl.requests[key][:0]
	for _, t := range rl.requests[key] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	remaining := rl.limit - len(valid)
	if remaining <= 0 {
		rl.requests[key] = valid
		reset := windowStart.Unix()
		if len(valid) > 0 {
			reset = valid[0].Add(rl.window).Unix()
		}
		return false, 0, reset
	}

	rl.requests[key] = append(valid, now)
	return true, remaining - 1, now.Add(rl.window).Unix()
}

func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)
	return func(ctx *gin.Context) {
		key := ctx.ClientIP()

		allowed, remaining, reset := limiter.allow(key)

		ctx.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		ctx.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))

		if !allowed {
			utils.RespondWithError(ctx, apperror.New(
				http.StatusTooManyRequests,
				429,
				"Too many requests. Please try again later.",
			))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
