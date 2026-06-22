package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

const (
	// MAX_BODY_SIZE is the maximum size of request and response body to log (64 KB)
	MAX_BODY_SIZE = 1 << 16 // 64 KB

	// NotLoggedResponse is the placeholder used when response body is not logged for successful requests
	NotLoggedResponse = "<not_log>"
)

// sensitiveKeys are field names that contain sensitive data and should be censored in logs
var sensitiveKeys = []string{
	"password", "api-key", "token", "access_token", "refresh_token",
	"ccv", "credit_card", "debit_card", "social_security_number",
	"ssn", "bank_account", "bank_account_number",
	"email", "phone", "address", "cvv",
	"secret", "otp", "totp", "mfa_code", "mfa_secret",
	"verification_code", "new_password", "old_password", "confirm_password",
	"session", "session_id", "sessionid", "sid",
}

// sensitiveHeaders are HTTP headers that contain sensitive information
var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"cookie":              true,
	"set-cookie":          true,
	"x-api-key":           true,
	"x-auth-token":        true,
	"proxy-authorization": true,
	"x-forwarded-for":     true,
	"x-real-ip":           true,
	"forwarded":           true,
	"true-client-ip":      true,
	"x-csrf-token":        true,
	"xsrf-token":          true,
	"x-xsrf-token":        true,
}

var marshalLogEntry = json.Marshal

// LogResponse defines the structure for logging HTTP requests and responses
type LogResponse struct {
	RequestID  string `json:"request_id,omitempty"`
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

// censorQueryParams censors sensitive query parameters (keeps first 4 chars)
func censorQueryParams(queryParams map[string][]string) map[string][]string {
	censored := make(map[string][]string, len(queryParams))
	for key, values := range queryParams {
		if containsIgnoreCase(sensitiveKeys, key) {
			masked := make([]string, len(values))
			for i, v := range values {
				masked[i] = utils.MaskWithPrefix(v, 4)
			}
			censored[key] = masked
		} else {
			censored[key] = values
		}
	}
	return censored
}

// containsIgnoreCase checks if a string matches any key in the list (case-insensitive)
func containsIgnoreCase(keys []string, target string) bool {
	targetLower := strings.ToLower(target)
	for _, k := range keys {
		if strings.ToLower(k) == targetLower {
			return true
		}
	}
	return false
}

// maskCookieValue parses a cookie string and masks sensitive values while keeping keys.
// Known sensitive cookie keys (session, token, auth) are masked.
// Technical attributes (Path, Domain, HttpOnly, Secure) are kept for debugging.
// Example: "session_id=abc123; Path=/" → "session_id=abc1*****; Path=/"
func maskCookieValue(value string) string {
	parts := strings.Split(value, ";")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		eqIdx := strings.Index(part, "=")
		if eqIdx > 0 {
			key := part[:eqIdx]
			if containsIgnoreCase(sensitiveKeys, key) {
				rawVal := part[eqIdx+1:]
				parts[i] = key + "=" + utils.MaskWithPrefix(rawVal, 4)
			} else {
				parts[i] = part // keep non-sensitive attributes, trimmed
			}
		} else {
			parts[i] = part // keep flag attributes (HttpOnly, Secure), trimmed
		}
	}
	return strings.Join(parts, "; ")
}

// filterSensitiveHeaders creates a copy of headers with sensitive values censored
func filterSensitiveHeaders(headers map[string][]string) map[string][]string {
	filtered := make(map[string][]string, len(headers))
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			if lowerKey == "authorization" && len(values) > 0 {
				parts := strings.SplitN(values[0], " ", 2)
				if len(parts) == 2 {
					// Show the scheme (e.g., "Bearer") and first 4 chars of token
					filtered[key] = []string{parts[0] + " " + utils.MaskWithPrefix(parts[1], 4)}
				} else {
					filtered[key] = []string{utils.MaskWithPrefix(values[0], 4)}
				}
			} else if lowerKey == "cookie" || lowerKey == "set-cookie" {
				masked := make([]string, len(values))
				for i, v := range values {
					masked[i] = maskCookieValue(v)
				}
				filtered[key] = masked
			} else {
				filtered[key] = []string{utils.MaskWithPrefix(values[0], 4)}
			}
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
			RequestID: GetRequestID(c),
			Method:    c.Request.Method,
			URL:       c.Request.URL.String(),
			Header:    filterSensitiveHeaders(c.Request.Header),
			Request:   censorQueryParams(c.Request.URL.Query()),
		}

		// Only log request body if method is POST or PUT, and limit to maxBodySize
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			var bodyBytes []byte
			if c.Request.Body != nil {
				var err error
				bodyBytes, err = io.ReadAll(io.LimitReader(c.Request.Body, MAX_BODY_SIZE))
				if err != nil {
					logger.WithField("request_id", logEntry.RequestID).Errorf("Failed to read request body: %v", err)
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

		statusCode := c.Writer.Status()
		logEntry.Latency = fmt.Sprintf("%d (ms)", time.Since(timeStart).Milliseconds())
		logEntry.StatusCode = fmt.Sprintf("%d", statusCode)

		// Only log response body for error status codes (>= 400)
		if statusCode >= 400 {
			respBodyBytes := responseBody.Bytes()
			if len(respBodyBytes) > MAX_BODY_SIZE {
				respBodyBytes = respBodyBytes[:MAX_BODY_SIZE]
			}

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
		} else {
			logEntry.Response = NotLoggedResponse
		}

		// Use goroutine to write log entry to avoid blocking
		go func(entry LogResponse, sc int) {
			fields := log.Fields{
				"request_id":  entry.RequestID,
				"method":      entry.Method,
				"url":         entry.URL,
				"status_code": entry.StatusCode,
				"latency":     entry.Latency,
				"header":      entry.Header,
				"request":     entry.Request,
				"response":    entry.Response,
			}
			switch {
			case sc >= 500:
				logger.WithFields(fields).Error("HTTP request completed")
			case sc >= 400:
				logger.WithFields(fields).Warn("HTTP request completed")
			default:
				logger.WithFields(fields).Info("HTTP request completed")
			}
		}(logEntry, statusCode)
	}
}
