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

	t.Run("ParseDateString_EmptyString", func(t *testing.T) {
		parsedTime, err := ParseDateStringYYYYMMDD("")
		assert.Error(t, err)
		assert.Nil(t, parsedTime)
	})

	t.Run("ParseDateString_InvalidCalendarDate", func(t *testing.T) {
		// 2023-02-29 does not exist
		parsedTime, err := ParseDateStringYYYYMMDD("2023-02-29")
		assert.Error(t, err)
		assert.Nil(t, parsedTime)
	})

	t.Run("ParseDateString_LeapYear", func(t *testing.T) {
		// 2024-02-29 is a valid leap day
		parsedTime, err := ParseDateStringYYYYMMDD("2024-02-29")
		assert.NoError(t, err)
		assert.Equal(t, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), *parsedTime)
	})

	t.Run("ParseDateString_InvalidMonth", func(t *testing.T) {
		parsedTime, err := ParseDateStringYYYYMMDD("2023-13-01")
		assert.Error(t, err)
		assert.Nil(t, parsedTime)
	})
}
