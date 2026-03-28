package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/liamhendricks/ghostman/pkg/parser"
)

// Executor sends HTTP requests using the provided HTTPDoer.
type Executor struct {
	Client HTTPDoer
}

// New creates a new Executor with the given HTTP client.
func New(client HTTPDoer) *Executor {
	return &Executor{Client: client}
}

// contentTypeForLang maps a BodyLang value to its corresponding Content-Type.
// Returns empty string when no Content-Type should be auto-set.
func contentTypeForLang(lang string) string {
	switch lang {
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	case "form":
		return "application/x-www-form-urlencoded"
	case "graphql":
		return "application/json"
	case "body", "":
		return ""
	default:
		return ""
	}
}

// Execute builds and sends an HTTP request from the given RequestSpec.
// Returns (Response, error) where error is non-nil for both transport errors
// and HTTP 4xx/5xx responses. Response.Body is always populated, even on 4xx/5xx.
func (e *Executor) Execute(ctx context.Context, spec parser.RequestSpec) (Response, error) {
	// 1. Build URL: spec.BaseURL + spec.Path
	rawURL := spec.BaseURL + spec.Path

	// 2. Append query params if present
	if len(spec.QueryParams) > 0 {
		u, err := url.Parse(rawURL)
		if err != nil {
			return Response{}, fmt.Errorf("invalid URL %q: %w", rawURL, err)
		}
		q := u.Query()
		for k, v := range spec.QueryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		rawURL = u.String()
	}

	// 3. Build body reader
	var bodyReader io.Reader
	if len(spec.Body) > 0 {
		bodyReader = bytes.NewReader(spec.Body)
	}

	// 4. Create request
	req, err := http.NewRequestWithContext(ctx, spec.Method, rawURL, bodyReader)
	if err != nil {
		return Response{}, fmt.Errorf("failed to build request: %w", err)
	}

	// 5. Set headers from spec
	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}

	// 6. Auto-set Content-Type from BodyLang if body present and Content-Type not already set
	if len(spec.Body) > 0 && req.Header.Get("Content-Type") == "" {
		if ct := contentTypeForLang(spec.BodyLang); ct != "" {
			req.Header.Set("Content-Type", ct)
		}
	}

	// 7. Record start time
	start := time.Now()

	// 8. Send request
	resp, err := e.Client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 9-10. Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// 11. Compute duration
	duration := time.Since(start)

	// 12. Build Response
	response := Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       bodyBytes,
		Duration:   duration,
		BodySize:   int64(len(bodyBytes)),
	}

	// 13. Return HTTPError for 4xx/5xx (body is still populated)
	if resp.StatusCode >= 400 {
		return response, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	// 14. Success
	return response, nil
}
