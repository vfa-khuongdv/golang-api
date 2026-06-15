package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCensorQueryParams(t *testing.T) {
	t.Run("sensitive key masked", func(t *testing.T) {
		input := map[string][]string{"token": {"abc123"}, "q": {"search"}}
		result := censorQueryParams(input)
		assert.Equal(t, "abc1*****", result["token"][0])
		assert.Equal(t, "search", result["q"][0])
	})

	t.Run("non-sensitive key unchanged", func(t *testing.T) {
		input := map[string][]string{"page": {"2"}, "limit": {"10"}}
		result := censorQueryParams(input)
		assert.Equal(t, "2", result["page"][0])
		assert.Equal(t, "10", result["limit"][0])
	})

	t.Run("empty params", func(t *testing.T) {
		result := censorQueryParams(map[string][]string{})
		assert.Empty(t, result)
	})
}

func TestContainsIgnoreCase(t *testing.T) {
	t.Run("match found", func(t *testing.T) {
		assert.True(t, containsIgnoreCase([]string{"Password", "Token"}, "password"))
		assert.True(t, containsIgnoreCase([]string{"Password", "Token"}, "TOKEN"))
		assert.True(t, containsIgnoreCase([]string{"api-key"}, "API-KEY"))
	})

	t.Run("no match", func(t *testing.T) {
		assert.False(t, containsIgnoreCase([]string{"Password", "Token"}, "username"))
		assert.False(t, containsIgnoreCase([]string{}, "anything"))
	})
}

func TestFilterSensitiveHeaders(t *testing.T) {
	t.Run("authorization with bearer", func(t *testing.T) {
		input := map[string][]string{"Authorization": {"Bearer eyJhbGci"}}
		result := filterSensitiveHeaders(input)
		assert.Equal(t, "Bearer eyJh*****", result["Authorization"][0])
	})

	t.Run("authorization without scheme", func(t *testing.T) {
		input := map[string][]string{"Authorization": {"token-abc"}}
		result := filterSensitiveHeaders(input)
		assert.Contains(t, result["Authorization"][0], "*****")
	})

	t.Run("cookie masked", func(t *testing.T) {
		input := map[string][]string{"Cookie": {"session_id=abc123"}}
		result := filterSensitiveHeaders(input)
		assert.Equal(t, "session_id=abc1*****", result["Cookie"][0])
	})

	t.Run("set-cookie masked", func(t *testing.T) {
		input := map[string][]string{"Set-Cookie": {"session_id=abc123; Path=/; HttpOnly"}}
		result := filterSensitiveHeaders(input)
		assert.Equal(t, "session_id=abc1*****; Path=/; HttpOnly", result["Set-Cookie"][0])
	})

	t.Run("non-sensitive header unchanged", func(t *testing.T) {
		input := map[string][]string{"Content-Type": {"application/json"}}
		result := filterSensitiveHeaders(input)
		assert.Equal(t, "application/json", result["Content-Type"][0])
	})
}
