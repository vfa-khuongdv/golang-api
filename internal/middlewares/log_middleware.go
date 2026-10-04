package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

// sensitiveKeys are field names that contain sensitive data and should be censored in logs.
// A "*word*" entry matches any field whose name contains the word (e.g. mail_password).
var sensitiveKeys = []string{
	"*password*", "*secret*", "*token*",
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

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write forwards to the client and keeps a copy for logging only for error
// responses (the only ones whose body is logged), capped at MAX_BODY_SIZE.
func (w *bodyWriter) Write(b []byte) (int, error) {
	if w.Status() >= 400 {
		if remaining := MAX_BODY_SIZE - w.body.Len(); remaining > 0 {
			w.body.Write(b[:min(len(b), remaining)])
		}
	}
	return w.ResponseWriter.Write(b)
}

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

func containsIgnoreCase(keys []string, target string) bool {
	for _, k := range keys {
		if utils.MatchesSensitiveKey(k, target) {
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
				parts[i] = part
			}
		} else {
			parts[i] = part
		}
	}
	return strings.Join(parts, "; ")
}

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
		// Load balancer probes hit this every few seconds on every task.
		if path := c.Request.URL.Path; path == "/healthz" || path == "/readyz" {
			c.Next()
			return
		}

		timeStart := time.Now()

		// The URL is logged with the same masked query, so a token passed in the
		// query string (e.g. ?token=...) never reaches the logs.
		query := censorQueryParams(c.Request.URL.Query())
		logURL := c.Request.URL.Path
		if len(query) > 0 {
			logURL += "?" + url.Values(query).Encode()
		}

		logEntry := LogResponse{
			RequestID: GetRequestID(c),
			Method:    c.Request.Method,
			URL:       logURL,
			Header:    filterSensitiveHeaders(c.Request.Header),
			Request:   query,
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
				// Only the first MAX_BODY_SIZE bytes are logged; the handler still reads the whole body.
				c.Request.Body = struct {
					io.Reader
					io.Closer
				}{io.MultiReader(bytes.NewReader(bodyBytes), c.Request.Body), c.Request.Body}
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

		// Response body capture is capped at MAX_BODY_SIZE and only allocated for errors
		responseBody := &bytes.Buffer{}
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

		// Written before the request returns: an entry written from a goroutine
		// can be lost on shutdown or interleave with later requests.
		fields := log.Fields{
			"request_id":  logEntry.RequestID,
			"method":      logEntry.Method,
			"url":         logEntry.URL,
			"status_code": logEntry.StatusCode,
			"latency":     logEntry.Latency,
			"header":      logEntry.Header,
			"request":     logEntry.Request,
			"response":    logEntry.Response,
		}
		switch {
		case statusCode >= 500:
			logger.WithFields(fields).Error("HTTP request completed")
		case statusCode >= 400:
			logger.WithFields(fields).Warn("HTTP request completed")
		default:
			logger.WithFields(fields).Info("HTTP request completed")
		}
	}
}
