package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateString(t *testing.T) {
	t.Run("ParseDateString_ValidDate", func(t *testing.T) {
		dateStr := "2023-10-15"
		parsedTime, err := ParseDateStringYYYYMMDD(dateStr)
		assert.NoError(t, err)
		expectedTime := time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedTime, *parsedTime)
	})

	t.Run("ParseDateString_InvalidDate", func(t *testing.T) {
		dateStr := "15-10-2023"
		parsedTime, err := ParseDateStringYYYYMMDD(dateStr)
		assert.Error(t, err)
		assert.Nil(t, parsedTime)
	})
}
