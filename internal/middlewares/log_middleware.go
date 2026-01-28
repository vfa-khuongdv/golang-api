package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

const (
	// MAX_BODY_SIZE is the maximum size of request and response body to log (64 KB)
	MAX_BODY_SIZE = 1 << 16 // 64 KB
)

// sensitiveKeys are field names that contain sensitive data and should be censored in logs
var sensitiveKeys = []string{
	"password", "api-key", "token", "access_token", "refresh_token",
	"ccv", "credit_card", "debit_card", "social_security_number",
	"ssn", "bank_account", "bank_account_number",
	"email", "phone", "address", "cvv",
}

// sensitiveHeaders are HTTP headers that contain sensitive information
var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"cookie":              true,
	"set-cookie":          true,
	"x-api-key":           true,
	"x-auth-token":        true,
	"proxy-authorization": true,
}

// LogResponse defines the structure for logging HTTP requests and responses
type LogResponse struct {
	Method     string `json:"method"`
	URL        string `json:"url"`
	Header     any    `json:"header"`
	Request    any    `json:"request,omitempty"`
	Response   any    `json:"response,omitempty"`
	Latency    string `json:"latency,omitempty"`
	StatusCode string `json:"status_code"`
}

// the bodyWriter is a custom ResponseWriter that captures the response body
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// filterSensitiveHeaders creates a copy of headers with sensitive values censored
func filterSensitiveHeaders(headers map[string][]string) map[string][]string {
	filtered := make(map[string][]string, len(headers))
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			filtered[key] = []string{"*****"}
		} else {
			filtered[key] = values
		}
	}
	return filtered
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeStart := time.Now()

		logEntry := LogResponse{
			Method:  c.Request.Method,
			URL:     c.Request.URL.String(),
			Header:  filterSensitiveHeaders(c.Request.Header),
			Request: c.Request.URL.Query(),
		}

		// Only log request body if method is POST or PUT, and limit to maxBodySize
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			var bodyBytes []byte
			if c.Request.Body != nil {
				var err error
				bodyBytes, err = io.ReadAll(io.LimitReader(c.Request.Body, MAX_BODY_SIZE))
				if err != nil {
					logger.Error("Failed to read request body:", err)
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			if strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
				var requestBody any
				if err := json.Unmarshal(bodyBytes, &requestBody); err == nil {
					requestBody = utils.CensorSensitiveData(requestBody, sensitiveKeys)
					logEntry.Request = requestBody
				} else {
					logEntry.Request = string(bodyBytes)
				}
			} else {
				logEntry.Request = string(bodyBytes)
			}
		}

		// Limit response body capture to MAX_BODY_SIZE
		responseBody := bytes.NewBuffer(make([]byte, 0, MAX_BODY_SIZE))
		c.Writer = &bodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}

		c.Next()

		logEntry.Latency = fmt.Sprintf("%d (ms)", time.Since(timeStart).Milliseconds())
		logEntry.StatusCode = fmt.Sprintf("%d", c.Writer.Status())

		// Limit response body to MAX_BODY_SIZE for logging
		respBodyBytes := responseBody.Bytes()
		if len(respBodyBytes) > MAX_BODY_SIZE {
			respBodyBytes = respBodyBytes[:MAX_BODY_SIZE]
		}

		// If response is JSON, unmarshal and censor sensitive data
		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			var responseBodyData any
			if err := json.Unmarshal(respBodyBytes, &responseBodyData); err == nil {
				responseBodyData = utils.CensorSensitiveData(responseBodyData, sensitiveKeys)
				logEntry.Response = responseBodyData
			} else {
				logEntry.Response = string(respBodyBytes)
			}
		} else {
			logEntry.Response = string(respBodyBytes)
		}

		// Use goroutine to write log entry to avoid blocking
		go func(entry LogResponse) {
			jsonData, err := json.Marshal(entry)
			if err != nil {
				logger.Error("Failed to marshal log entry:", err)
				return
			}
			logger.Info(string(jsonData))
		}(logEntry)
	}
}
