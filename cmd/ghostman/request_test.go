package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testRequestSetup creates a temp directory with a .ghostman collection structure
// pointing to the provided httptest server URL. Returns the temp dir path and a
// cleanup function.
func testRequestSetup(t *testing.T, server *httptest.Server, extraFiles map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	ghostmanDir := filepath.Join(dir, ".ghostman")
	authDir := filepath.Join(ghostmanDir, "auth")
	envDir := filepath.Join(ghostmanDir, "env")

	for _, d := range []string{authDir, envDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Default login request file
	loginMD := fmt.Sprintf(`---
ghostman_version: 1
method: GET
base_url: %s
path: /hello
---
`, server.URL)
	if err := os.WriteFile(filepath.Join(authDir, "login.md"), []byte(loginMD), 0o644); err != nil {
		t.Fatal(err)
	}

	// Default test env file
	testEnvMD := `---
ghostman_version: 1
---

` + "```" + `env
token=test-token-123
` + "```" + `
`
	if err := os.WriteFile(filepath.Join(envDir, "test.md"), []byte(testEnvMD), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write any extra files
	for relPath, content := range extraFiles {
		fullPath := filepath.Join(ghostmanDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// executeRequest runs the root command with the given args from workDir,
// returning stdout, stderr, and the error.
func executeRequest(t *testing.T, workDir string, args []string) (stdout, stderr string, err error) {
	t.Helper()

	origDir, _ := os.Getwd()
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)

	execErr := root.Execute()
	return outBuf.String(), errBuf.String(), execErr
}

func TestRequestCommand_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	}))
	defer server.Close()

	dir := testRequestSetup(t, server, nil)

	stdout, stderr, err := executeRequest(t, dir, []string{"request", "auth/login"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Body should be on stdout
	if !strings.Contains(stdout, `{"status":"ok"}`) {
		t.Errorf("expected body on stdout, got: %q", stdout)
	}

	// Metadata should be on stderr
	if !strings.Contains(stderr, "200") {
		t.Errorf("expected status code on stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "bytes") {
		t.Errorf("expected byte count on stderr, got: %q", stderr)
	}
}

func TestRequestCommand_4xxExitsNonZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`)) //nolint:errcheck
	}))
	defer server.Close()

	dir := testRequestSetup(t, server, nil)

	stdout, stderr, err := executeRequest(t, dir, []string{"request", "auth/login"})

	// Should exit non-zero
	if err == nil {
		t.Fatal("expected non-zero exit on 4xx, got nil error")
	}

	// Body should still be on stdout
	if !strings.Contains(stdout, `{"error":"bad request"}`) {
		t.Errorf("expected body on stdout even on 4xx, got: %q", stdout)
	}

	// Metadata should still appear on stderr
	if !strings.Contains(stderr, "400") {
		t.Errorf("expected 400 on stderr, got: %q", stderr)
	}
}

func TestRequestCommand_DryRun(t *testing.T) {
	hitCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dir := testRequestSetup(t, server, nil)

	stdout, _, err := executeRequest(t, dir, []string{"request", "--dry-run", "auth/login"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// No HTTP call should have been made
	if hitCount != 0 {
		t.Errorf("expected no HTTP calls in dry-run mode, got %d", hitCount)
	}

	// Dry run output should describe the request
	if !strings.Contains(stdout, "DRY RUN") {
		t.Errorf("expected DRY RUN header in stdout, got: %q", stdout)
	}
	if !strings.Contains(stdout, "METHOD") {
		t.Errorf("expected METHOD in dry-run output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "URL") {
		t.Errorf("expected URL in dry-run output, got: %q", stdout)
	}
}

func TestRequestCommand_MissingEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dir := t.TempDir()
	ghostmanDir := filepath.Join(dir, ".ghostman")
	authDir := filepath.Join(ghostmanDir, "auth")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Request with {{env:*}} reference but no --env flag
	loginMD := fmt.Sprintf(`---
ghostman_version: 1
method: GET
base_url: %s
path: /hello
headers:
  Authorization: "Bearer {{env:token}}"
---
`, server.URL)
	if err := os.WriteFile(filepath.Join(authDir, "login.md"), []byte(loginMD), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := executeRequest(t, dir, []string{"request", "auth/login"})
	if err == nil {
		t.Fatal("expected error when env refs exist but --env not provided")
	}
	if !strings.Contains(err.Error(), "--env is required") {
		t.Errorf("expected '--env is required' in error, got: %v", err)
	}
}

func TestRequestCommand_OutputSeparation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response-body-content")) //nolint:errcheck
	}))
	defer server.Close()

	dir := testRequestSetup(t, server, nil)

	stdout, stderr, err := executeRequest(t, dir, []string{"request", "auth/login"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Stdout should contain ONLY the response body
	if !strings.Contains(stdout, "response-body-content") {
		t.Errorf("expected response body on stdout, got: %q", stdout)
	}
	// Stdout must NOT contain status code or timing
	if strings.Contains(stdout, "200") {
		t.Errorf("stdout must not contain status code, got: %q", stdout)
	}

	// Stderr should contain metadata (status code, timing, bytes)
	if !strings.Contains(stderr, "200") {
		t.Errorf("expected 200 on stderr, got: %q", stderr)
	}
	if strings.Contains(stderr, "response-body-content") {
		t.Errorf("stderr must not contain response body, got: %q", stderr)
	}
}

func TestRequestCommand_CertWithoutKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	certOnlyMD := fmt.Sprintf(`---
ghostman_version: 1
method: GET
base_url: %s
path: /secure
cert: /some/client.crt
---
`, server.URL)

	dir := testRequestSetup(t, server, map[string]string{
		"auth/cert-only.md": certOnlyMD,
	})

	_, _, err := executeRequest(t, dir, []string{"request", "auth/cert-only"})
	if err == nil {
		t.Fatal("expected error when cert is set without key")
	}
	if !strings.Contains(err.Error(), "cert and key must both be set") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequestCommand_KeyWithoutCert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	keyOnlyMD := fmt.Sprintf(`---
ghostman_version: 1
method: GET
base_url: %s
path: /secure
key: /some/client.key
---
`, server.URL)

	dir := testRequestSetup(t, server, map[string]string{
		"auth/key-only.md": keyOnlyMD,
	})

	_, _, err := executeRequest(t, dir, []string{"request", "auth/key-only"})
	if err == nil {
		t.Fatal("expected error when key is set without cert")
	}
	if !strings.Contains(err.Error(), "cert and key must both be set") {
		t.Errorf("unexpected error: %v", err)
	}
}
