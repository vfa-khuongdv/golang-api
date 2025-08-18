package services_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

func setupTestRedis(t *testing.T) (*services.RedisService, func()) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	service := services.NewRedisService(client)

	// teardown function
	return service, func() {
		_ = client.Close()
		s.Close()
	}
}

func TestRedisService_Set(t *testing.T) {
	svc, teardown := setupTestRedis(t)
	defer teardown()

	err := svc.Set("key", "value", time.Minute)
	assert.NoError(t, err)
}

func TestRedisService_Get(t *testing.T) {
	svc, teardown := setupTestRedis(t)
	defer teardown()

	// First set the value so Get can retrieve it
	err := svc.Set("key", "value", time.Minute)
	assert.NoError(t, err)

	val, err := svc.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestRedisService_Delete(t *testing.T) {
	svc, teardown := setupTestRedis(t)
	defer teardown()

	_ = svc.Set("key", "value", time.Minute)

	err := svc.Delete("key")
	assert.NoError(t, err)

	val, err := svc.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestRedisService_Exists(t *testing.T) {
	svc, teardown := setupTestRedis(t)
	defer teardown()

	_ = svc.Set("key", "value", time.Minute)

	exists, err := svc.Exists("key")
	assert.NoError(t, err)
	assert.True(t, exists)

	_ = svc.Delete("key")

	exists, err = svc.Exists("key")
	assert.NoError(t, err)
	assert.False(t, exists)
}
