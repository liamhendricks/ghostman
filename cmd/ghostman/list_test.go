package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// executeList runs the root command with "list" from workDir,
// returning stdout, stderr, and the error.
func executeList(t *testing.T, workDir string) (stdout, stderr string, err error) {
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
	root.SetArgs([]string{"list"})

	execErr := root.Execute()
	return outBuf.String(), errBuf.String(), execErr
}

func TestListCommand_Output(t *testing.T) {
	dir := t.TempDir()
	ghostmanDir := filepath.Join(dir, ".ghostman")

	// Create auth/login.md and users/list.md
	for _, path := range []string{"auth/login.md", "users/list.md"} {
		fullPath := filepath.Join(ghostmanDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("# placeholder"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stdout, _, err := executeList(t, dir)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	found := make(map[string]bool)
	for _, line := range lines {
		found[line] = true
	}

	if !found["auth/login"] {
		t.Errorf("expected 'auth/login' in output, got lines: %v", lines)
	}
	if !found["users/list"] {
		t.Errorf("expected 'users/list' in output, got lines: %v", lines)
	}

	// Verify one per line (no concatenation)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 1 {
			t.Errorf("expected one path per line, got: %q", line)
		}
	}
}

func TestListCommand_NestedCollections(t *testing.T) {
	dir := t.TempDir()
	ghostmanDir := filepath.Join(dir, ".ghostman")

	// Create a three-level deep request: collection/resource/action.md
	for _, path := range []string{"api/users/list.md", "api/users/create.md", "api/posts/get.md"} {
		fullPath := filepath.Join(ghostmanDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("# placeholder"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stdout, _, err := executeList(t, dir)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	found := make(map[string]bool)
	for _, line := range lines {
		found[line] = true
	}

	for _, want := range []string{"api/users/list", "api/users/create", "api/posts/get"} {
		if !found[want] {
			t.Errorf("expected %q in output, got lines: %v", want, lines)
		}
	}
}

func TestListCommand_ExcludesEnvAndVars(t *testing.T) {
	dir := t.TempDir()
	ghostmanDir := filepath.Join(dir, ".ghostman")

	// Create files that should be excluded
	for _, path := range []string{
		"env/staging.md",  // in env/ subdir — excluded
		"auth/vars.md",    // named vars.md — excluded
		"auth/login.md",   // regular request — included
	} {
		fullPath := filepath.Join(ghostmanDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("# placeholder"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stdout, _, err := executeList(t, dir)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if strings.Contains(stdout, "env/staging") {
		t.Errorf("env/staging.md should be excluded from list output, got: %q", stdout)
	}
	if strings.Contains(stdout, "vars") {
		t.Errorf("vars.md should be excluded from list output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "auth/login") {
		t.Errorf("auth/login should appear in output, got: %q", stdout)
	}
}

func TestListCommand_EmptyCollection(t *testing.T) {
	dir := t.TempDir()
	ghostmanDir := filepath.Join(dir, ".ghostman")

	if err := os.MkdirAll(ghostmanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeList(t, dir)
	if err != nil {
		t.Fatalf("expected success on empty collection, got error: %v", err)
	}

	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected empty output for empty collection, got: %q", stdout)
	}
}
