package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

// TestGenerateRandomString checks random string generation
func TestGenerateRandomString(t *testing.T) {
	t.Run("checks that the generated string is of the correct length", func(t *testing.T) {
		length := 10
		randomStr := utils.GenerateRandomString(length)

		require.Equal(t, length, len(randomStr), "Expected length %d, but got %d", length, len(randomStr))
	})

	t.Run("checks that consecutive calls return different strings", func(t *testing.T) {
		str1 := utils.GenerateRandomString(10)
		str2 := utils.GenerateRandomString(10)

		require.NotEqual(t, str1, str2, "Expected different strings but got the same: %s", str1)
	})

	t.Run("ensures the result only contains allowed characters", func(t *testing.T) {
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		allowed := map[rune]bool{}
		for _, ch := range charset {
			allowed[ch] = true
		}

		randomStr := utils.GenerateRandomString(100)
		for _, ch := range randomStr {
			require.True(t, allowed[ch], "Generated string contains invalid character: %c", ch)
		}
	})
}

func TestStringToPtr(t *testing.T) {
	t.Run("returns pointer when string is non-empty", func(t *testing.T) {
		input := "hello"
		ptr := utils.StringToPtr(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})

	t.Run("returns nil when string is empty", func(t *testing.T) {
		ptr := utils.StringToPtr("")
		assert.Nil(t, ptr)
	})
}

func TestIntToPtr(t *testing.T) {
	t.Run("returns pointer to int", func(t *testing.T) {
		input := 42
		ptr := utils.IntToPtr(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})

	t.Run("returns pointer to float64", func(t *testing.T) {
		input := 3.14
		ptr := utils.IntToPtr(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})

	t.Run("returns pointer to bool", func(t *testing.T) {
		input := true
		ptr := utils.IntToPtr(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})

	t.Run("returns pointer to string", func(t *testing.T) {
		input := "test"
		ptr := utils.IntToPtr(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})
}
