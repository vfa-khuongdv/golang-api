package services

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisService creates and initializes a new Redis service connection
// Parameters:
//   - address: Redis server address in format "host:port"
//   - password: Redis server password (empty string if no password required)
//   - db: Redis database number to connect to
//
// Returns:
//   - *RedisService: Pointer to initialized Redis service
//   - Panics if connection fails
func NewRedisService(address, password string, db int) *RedisService {

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // leave empty if no password
		DB:       db,       // default DB
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()

	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	fmt.Printf("Redis run at %s\n", address)

	return &RedisService{
		client: client,
		ctx:    ctx,
	}

}

// Set stores a key-value pair in Redis with no expiration time
// Parameters:
//   - key: the key to store the value under
//   - value: the value to store (can be any type that Redis supports)
//
// Returns:
//   - error: nil if successful, otherwise contains the error message
func (r *RedisService) Set(key string, value interface{}) error {
	return r.client.Set(r.ctx, key, value, 0).Err()
}

// Get retrieves a value from Redis by its key
// Parameters:
//   - key: the key to look up in Redis
//
// Returns:
//   - string: the value stored at the key, or empty string if key doesn't exist
//   - error: nil if successful or key not found, otherwise contains error message
func (r *RedisService) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()

	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, err
}

// Delete removes a key-value pair from Redis
// Parameters:
//   - key: the key to delete from Redis
//
// Returns:
//   - error: nil if successful, otherwise contains the error message
func (r *RedisService) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists in Redis
// Parameters:
//   - key: the key to check in Redis
//
// Returns:
//   - bool: true if key exists, false if it doesn't
//   - error: nil if successful, otherwise contains the error message
func (r *RedisService) Exists(key string) (bool, error) {
	count, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetList retrieves all elements from a Redis list
// Parameters:
//   - listName: the name of the list to retrieve from Redis
//
// Returns:
//   - []string: slice containing all elements in the list
//   - error: nil if successful, otherwise contains the error message
func (e *RedisService) GetList(listName string) ([]string, error) {
	return e.client.LRange(e.ctx, listName, 0, -1).Result()
}

// Close closes the Redis client connection
// Returns:
//   - error: nil if successful, otherwise contains the error message
func (r *RedisService) Close() error {
	return r.client.Close()
}
