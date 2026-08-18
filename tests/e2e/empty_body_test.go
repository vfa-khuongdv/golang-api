package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// Regression: empty/whitespace-only request bodies are rejected centrally by
// TranslateValidationErrors (io.EOF -> ErrEmptyData) instead of a middleware,
// so a body-less POST like /logout still works while body-requiring endpoints
// return a clean 400.
func TestEmptyRequestBodyRejected(t *testing.T) {
	router, _ := setupTestRouter()

	emptyBodies := []struct {
		name string
		body []byte
	}{
		{name: "no body", body: nil},
		{name: "empty string", body: []byte("")},
		{name: "whitespace only", body: []byte("   \n\t")},
	}

	for _, tc := range emptyBodies {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errResp ErrorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
			assert.Equal(t, apperror.ErrEmptyData, errResp.Code)
			assert.Equal(t, "Request body cannot be empty", errResp.Message)
		})
	}
}
