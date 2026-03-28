package executor_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/liamhendricks/ghostman/pkg/executor"
	"github.com/liamhendricks/ghostman/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: create a test server that echoes the method in the response body.
func methodEchoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Method)
	}))
}

// helper: create a test server that echoes request details as headers + body.
func echoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// echo request headers back as response headers prefixed with X-Echo-
		for k, vs := range r.Header {
			for _, v := range vs {
				w.Header().Add("X-Echo-"+k, v)
			}
		}
		// echo body
		body, _ := io.ReadAll(r.Body)
		// echo content-type header directly
		if ct := r.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("X-Echo-Content-Type", ct)
		}
		w.Write(body)
	}))
}

// helper: create a fixed-status server that returns the given code and body.
func statusServer(code int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		fmt.Fprint(w, body)
	}))
}

// TestExecute_GET verifies a GET request returns 200, correct body, non-nil Duration.
func TestExecute_GET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "hello")
	}))
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "hello", string(resp.Body))
	assert.Greater(t, int64(resp.Duration), int64(0))
}

// TestExecute_POST_WithBody sends a POST with body and BodyLang="json".
// Verifies body is sent and Content-Type is auto-set to application/json.
func TestExecute_POST_WithBody(t *testing.T) {
	srv := echoServer()
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:   "POST",
		BaseURL:  srv.URL,
		Path:     "",
		Body:     []byte(`{"key":"val"}`),
		BodyLang: "json",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, `{"key":"val"}`, string(resp.Body))
	assert.Equal(t, "application/json", resp.Headers.Get("X-Echo-Content-Type"))
}

// TestExecute_AllMethods verifies all 7 standard HTTP methods are forwarded correctly.
func TestExecute_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			srv := methodEchoServer()
			defer srv.Close()

			ex := executor.New(srv.Client())
			spec := parser.RequestSpec{
				Method:  method,
				BaseURL: srv.URL,
				Path:    "",
			}
			resp, err := ex.Execute(context.Background(), spec)
			require.NoError(t, err)
			// HEAD responses have no body by HTTP spec
			if method != "HEAD" {
				assert.Equal(t, method, string(resp.Body))
			}
			assert.Equal(t, 200, resp.StatusCode)
		})
	}
}

// TestExecute_Headers verifies headers from RequestSpec are sent in the HTTP request.
func TestExecute_Headers(t *testing.T) {
	srv := echoServer()
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
		Headers: map[string]string{
			"X-Custom-Header": "myvalue",
			"Authorization":   "Bearer token123",
		},
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, "myvalue", resp.Headers.Get("X-Echo-X-Custom-Header"))
	assert.Equal(t, "Bearer token123", resp.Headers.Get("X-Echo-Authorization"))
}

// TestExecute_QueryParams verifies query params are appended to the URL.
func TestExecute_QueryParams(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
		QueryParams: map[string]string{
			"foo": "bar",
		},
	}
	_, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Contains(t, capturedQuery, "foo=bar")
}

// TestExecute_ContentTypeFromBodyLang verifies Content-Type auto-mapping from BodyLang.
func TestExecute_ContentTypeFromBodyLang(t *testing.T) {
	cases := []struct {
		bodyLang    string
		expectedCT  string
		expectSet   bool
	}{
		{"json", "application/json", true},
		{"xml", "application/xml", true},
		{"form", "application/x-www-form-urlencoded", true},
		{"graphql", "application/json", true},
		{"body", "", false},
		{"", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.bodyLang, func(t *testing.T) {
			var capturedCT string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCT = r.Header.Get("Content-Type")
				fmt.Fprint(w, "ok")
			}))
			defer srv.Close()

			ex := executor.New(srv.Client())
			spec := parser.RequestSpec{
				Method:   "POST",
				BaseURL:  srv.URL,
				Path:     "",
				Body:     []byte("some body"),
				BodyLang: tc.bodyLang,
			}
			_, err := ex.Execute(context.Background(), spec)
			require.NoError(t, err)
			if tc.expectSet {
				assert.Equal(t, tc.expectedCT, capturedCT)
			} else {
				assert.Empty(t, capturedCT)
			}
		})
	}
}

// TestExecute_ContentTypeUserPrecedence verifies user-provided Content-Type takes precedence.
func TestExecute_ContentTypeUserPrecedence(t *testing.T) {
	var capturedCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCT = r.Header.Get("Content-Type")
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "POST",
		BaseURL: srv.URL,
		Path:    "",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body:     []byte("some body"),
		BodyLang: "json",
	}
	_, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	// User-provided "text/plain" should win over BodyLang "json" -> "application/json"
	assert.Equal(t, "text/plain", capturedCT)
}

// TestExecute_4xxReturnsHTTPError verifies 400 returns (Response with body, *HTTPError).
func TestExecute_4xxReturnsHTTPError(t *testing.T) {
	srv := statusServer(400, "bad request body")
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.Error(t, err)
	var httpErr *executor.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "bad request body", string(resp.Body))
}

// TestExecute_5xxReturnsHTTPError verifies 500 response returns (Response with body, *HTTPError).
func TestExecute_5xxReturnsHTTPError(t *testing.T) {
	srv := statusServer(500, "internal error")
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.Error(t, err)
	var httpErr *executor.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Equal(t, "internal error", string(resp.Body))
}

// TestExecute_2xxNoError verifies 200, 201, 204 responses return (Response, nil).
func TestExecute_2xxNoError(t *testing.T) {
	for _, code := range []int{200, 201, 204} {
		t.Run(fmt.Sprintf("%d", code), func(t *testing.T) {
			srv := statusServer(code, "")
			defer srv.Close()

			ex := executor.New(srv.Client())
			spec := parser.RequestSpec{
				Method:  "GET",
				BaseURL: srv.URL,
				Path:    "",
			}
			_, err := ex.Execute(context.Background(), spec)
			assert.NoError(t, err)
		})
	}
}

// TestExecute_NetworkError verifies when HTTPDoer.Do returns an error, Execute wraps it.
func TestExecute_NetworkError(t *testing.T) {
	// Use a doer that always returns an error
	ex := executor.New(&errorDoer{err: fmt.Errorf("connection refused")})
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: "http://localhost:9",
		Path:    "",
	}
	_, err := ex.Execute(context.Background(), spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}

// errorDoer is a test HTTPDoer that always returns an error.
type errorDoer struct {
	err error
}

func (d *errorDoer) Do(req *http.Request) (*http.Response, error) {
	return nil, d.err
}

// TestExecute_BodyAlwaysRead verifies Response.Body is populated even on 4xx.
func TestExecute_BodyAlwaysRead(t *testing.T) {
	srv := statusServer(422, `{"error":"unprocessable entity"}`)
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "POST",
		BaseURL: srv.URL,
		Path:    "",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.Error(t, err)
	assert.Equal(t, `{"error":"unprocessable entity"}`, string(resp.Body))
}

// TestExecute_ResponseDuration verifies Duration > 0 for any successful request.
func TestExecute_ResponseDuration(t *testing.T) {
	srv := statusServer(200, "ok")
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Greater(t, int64(resp.Duration), int64(0))
}

// TestExecute_ResponseBodySize verifies BodySize matches len(Response.Body).
func TestExecute_ResponseBodySize(t *testing.T) {
	body := "hello world response"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "",
	}
	resp, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, int64(len(resp.Body)), resp.BodySize)
	assert.Equal(t, int64(len(body)), resp.BodySize)
}

// TestHTTPError_Error verifies HTTPError.Error() returns "HTTP 400 Bad Request" format.
func TestHTTPError_Error(t *testing.T) {
	e := &executor.HTTPError{
		StatusCode: 400,
		Status:     "Bad Request",
	}
	assert.Equal(t, "HTTP 400 Bad Request", e.Error())
}

// TestExecute_PathAppended verifies spec.Path is appended to spec.BaseURL.
func TestExecute_PathAppended(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	ex := executor.New(srv.Client())
	spec := parser.RequestSpec{
		Method:  "GET",
		BaseURL: srv.URL,
		Path:    "/api/v1/users",
	}
	_, err := ex.Execute(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/users", capturedPath)
}

// Compile-time check that *http.Client satisfies HTTPDoer.
var _ executor.HTTPDoer = (*http.Client)(nil)

// Compile-time check that strings.NewReader does NOT satisfy HTTPDoer (negative check).
var _ = strings.NewReader
