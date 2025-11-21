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

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeStart := time.Now()

		// List sensitive keys to censor in logs
		sensitiveKeys := []string{
			"password", "api-key", "token", "access_token", "refresh_token",
			"ccv", "credit_card", "debit_card", "social_security_number",
			"ssn", "bank_account", "bank_account_number",
			"email", "phone", "address", "cvv",
		}

		logEntry := LogResponse{
			Method:  c.Request.Method,
			URL:     c.Request.URL.String(),
			Header:  c.Request.Header,
			Request: c.Request.URL.Query(),
		}

		// Only log request body if method is POST or PUT, and limit to 64KB
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			const maxBodySize = 1 << 16 // 64 KB
			bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

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

		responseBody := &bytes.Buffer{}
		c.Writer = &bodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}

		c.Next()

		logEntry.Latency = fmt.Sprintf("%d (ms)", time.Since(timeStart).Milliseconds())
		logEntry.StatusCode = fmt.Sprintf("%d", c.Writer.Status())

		// If response is JSON, unmarshal and censor sensitive data
		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			var responseBodyData any
			if err := json.Unmarshal(responseBody.Bytes(), &responseBodyData); err == nil {
				responseBodyData = utils.CensorSensitiveData(responseBodyData, sensitiveKeys)
				logEntry.Response = responseBodyData
			} else {
				logEntry.Response = responseBody.String()
			}
		} else {
			logEntry.Response = responseBody.String()
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
