package executor

import (
	"fmt"
	"net/http"
	"time"
)

// Response holds the result of an HTTP request execution.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
	BodySize   int64
}

// HTTPError represents an HTTP-level error (4xx or 5xx response).
type HTTPError struct {
	StatusCode int
	Status     string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d %s", e.StatusCode, e.Status)
}

// HTTPDoer is the interface for sending HTTP requests.
// *http.Client satisfies this interface, enabling injection of httptest servers.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
