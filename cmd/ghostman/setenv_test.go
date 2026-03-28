package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetEnvCommand_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "local.yaml")

	// Create the env file with existing content
	if err := os.WriteFile(envFile, []byte("EXISTING: keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	jsonInput := `{"data":{"token":"abc123"}}`
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(jsonInput))
	root.SetArgs([]string{"set_env", "--env-file", envFile, "token", ".data.token"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Verify pass-through: original JSON on stdout
	out := outBuf.String()
	if !strings.Contains(out, `"data"`) {
		t.Errorf("expected original JSON on stdout (pass-through), got: %q", out)
	}

	// Verify the .env file was written
	content, readErr := os.ReadFile(envFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	fileContent := string(content)

	if !strings.Contains(fileContent, "token") {
		t.Errorf("expected 'token' key in env file, got: %q", fileContent)
	}
	if !strings.Contains(fileContent, "abc123") {
		t.Errorf("expected 'abc123' value in env file, got: %q", fileContent)
	}

	// Verify existing key is preserved
	if !strings.Contains(fileContent, "EXISTING") {
		t.Errorf("expected 'EXISTING' key to be preserved in env file, got: %q", fileContent)
	}
}

func TestSetEnvCommand_MissingPath(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "local.yaml")
	if err := os.WriteFile(envFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"data":{}}`))
	root.SetArgs([]string{"set_env", "--env-file", envFile, "token", ".missing"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing JSON path, got nil")
	}
}

func TestSetEnvCommand_MissingEnvFileFlag(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"x":1}`))
	root.SetArgs([]string{"set_env", "token", ".x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --env-file flag is missing, got nil")
	}
}

func TestSetEnvCommand_NonExistentEnvFile(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"x":"val"}`))
	root.SetArgs([]string{"set_env", "--env-file", "/nonexistent/path/.env", "mykey", ".x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent env file, got nil")
	}

	if !strings.Contains(err.Error(), "env file not found") {
		t.Errorf("expected 'env file not found' in error, got: %v", err)
	}
}
