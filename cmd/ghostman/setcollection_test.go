package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetCollectionCommand_WritesToVarsFile(t *testing.T) {
	dir := t.TempDir()
	// Create .ghostman/auth under tempdir
	collectionDir := filepath.Join(dir, ".ghostman", "auth")
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change cwd to tempdir so the command resolves .ghostman/auth correctly
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	jsonInput := `{"data":{"token":"abc123"}}`
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(jsonInput))
	root.SetArgs([]string{"--col", "auth", "set_collection", ".data.token"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Verify vars.yaml written with correct key and value
	varsPath := filepath.Join(collectionDir, "vars.yaml")
	content, readErr := os.ReadFile(varsPath)
	if readErr != nil {
		t.Fatalf("expected vars.yaml to be created, error: %v", readErr)
	}
	s := string(content)
	if !strings.Contains(s, "token") {
		t.Errorf("expected 'token' key in vars.yaml, got: %q", s)
	}
	if !strings.Contains(s, "abc123") {
		t.Errorf("expected 'abc123' value in vars.yaml, got: %q", s)
	}
}

func TestSetCollectionCommand_CreatesVarsFile(t *testing.T) {
	dir := t.TempDir()
	// Do NOT pre-create .ghostman/auth — command should create it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	jsonInput := `{"data":{"token":"xyz"}}`
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(jsonInput))
	root.SetArgs([]string{"--col", "auth", "set_collection", ".data.token"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	varsPath := filepath.Join(dir, ".ghostman", "auth", "vars.yaml")
	if _, err := os.Stat(varsPath); os.IsNotExist(err) {
		t.Fatal("expected vars.yaml to be created by set_collection")
	}
}

func TestSetCollectionCommand_MissingPath(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"data":{}}`))
	root.SetArgs([]string{"--col", "auth", "set_collection", ".missing"})

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for missing JSON path, got nil")
	}
	if !strings.Contains(err.Error(), "path not found") {
		t.Errorf("expected 'path not found' in error, got: %v", err)
	}
}

func TestSetCollectionCommand_MissingColFlag(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"data":{"token":"abc"}}`))
	root.SetArgs([]string{"set_collection", ".data.token"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --col flag is missing, got nil")
	}
	if !strings.Contains(err.Error(), "--col") {
		t.Errorf("expected '--col' in error message, got: %v", err)
	}
}

func TestSetCollectionCommand_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	collectionDir := filepath.Join(dir, ".ghostman", "auth")
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Pre-create vars.yaml with existing key
	varsPath := filepath.Join(collectionDir, "vars.yaml")
	if err := os.WriteFile(varsPath, []byte("base_url: http://example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"data":{"token":"tok123"}}`))
	root.SetArgs([]string{"--col", "auth", "set_collection", ".data.token"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	content, err := os.ReadFile(varsPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "base_url") {
		t.Errorf("expected 'base_url' to be preserved, got: %q", s)
	}
	if !strings.Contains(s, "http://example.com") {
		t.Errorf("expected 'http://example.com' to be preserved, got: %q", s)
	}
	if !strings.Contains(s, "token") {
		t.Errorf("expected 'token' key to be added, got: %q", s)
	}
	if !strings.Contains(s, "tok123") {
		t.Errorf("expected 'tok123' value to be added, got: %q", s)
	}
}

func TestSetCollectionCommand_PassThrough(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	jsonInput := `{"data":{"token":"passme"}}`
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(jsonInput))
	root.SetArgs([]string{"--col", "auth", "set_collection", ".data.token"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, `"data"`) {
		t.Errorf("expected original JSON on stdout (pass-through), got: %q", out)
	}
	if !strings.Contains(out, "passme") {
		t.Errorf("expected 'passme' value in stdout pass-through, got: %q", out)
	}
}
