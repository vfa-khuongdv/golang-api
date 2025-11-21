package utils_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

// helper to create a test context with query parameters
func createTestContextWithQuery(params map[string]string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	req, _ := http.NewRequest("GET", "/?"+query.Encode(), nil)
	c.Request = req
	return c
}

// TestCalculateTotalPages tests the pagination logic
func TestPaging(t *testing.T) {
	t.Run("CalculateTotalPages", func(t *testing.T) {
		tests := []struct {
			totalRows int64
			limit     int
			expected  int
		}{
			{totalRows: 100, limit: 10, expected: 10},
			{totalRows: 101, limit: 10, expected: 11},
			{totalRows: 0, limit: 10, expected: 0},
			{totalRows: 10, limit: 0, expected: 0},
			{totalRows: 10, limit: -1, expected: 0},
		}

		for _, tt := range tests {
			result := utils.CalculateTotalPages(tt.totalRows, tt.limit)
			assert.Equal(t, tt.expected, result)
		}
	})

	// TestParsePageAndLimit tests query parsing for pagination
	t.Run("ParsePageAndLimit", func(t *testing.T) {
		tests := []struct {
			queryParams   map[string]string
			expectedPage  int
			expectedLimit int
		}{
			{map[string]string{"page": "2", "limit": "20"}, 2, 20},
			{map[string]string{"page": "0", "limit": "0"}, 1, constants.LIMIT},
			{map[string]string{"page": "-1", "limit": "-10"}, 1, constants.LIMIT},
			{map[string]string{"page": "abc", "limit": "xyz"}, 1, constants.LIMIT},
			{map[string]string{}, 1, constants.LIMIT},
		}

		for _, tt := range tests {
			c := createTestContextWithQuery(tt.queryParams)
			page, limit := utils.ParsePageAndLimit(c)

			assert.Equal(t, tt.expectedPage, page)
			assert.Equal(t, tt.expectedLimit, limit)
		}
	})
}
