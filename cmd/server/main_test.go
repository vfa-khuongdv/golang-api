package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer(t *testing.T) {
	handler := http.NewServeMux()

	srv := newHTTPServer(":3000", handler)

	assert.Equal(t, ":3000", srv.Addr)
	assert.Equal(t, http.Handler(handler), srv.Handler)
	// Without these a slow client can hold a connection forever (slowloris), and
	// an idle timeout shorter than the ALB's 60s makes the ALB reuse closed
	// connections and return 502.
	assert.Equal(t, 10*time.Second, srv.ReadHeaderTimeout)
	assert.Greater(t, srv.IdleTimeout, 60*time.Second)
}
