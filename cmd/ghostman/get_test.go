package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGetCommand_ExtractsValue(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"data":{"token":"abc123"}}`))
	root.SetArgs([]string{"get", ".data.token"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, "abc123") {
		t.Errorf("expected 'abc123' on stdout, got: %q", out)
	}
}

func TestGetCommand_MissingPath(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"data":{}}`))
	root.SetArgs([]string{"get", ".missing"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}

	if !strings.Contains(err.Error(), "path not found") {
		t.Errorf("expected 'path not found' in error, got: %v", err)
	}
}

func TestGetCommand_NoArgs(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"x":1}`))
	root.SetArgs([]string{"get"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided, got nil")
	}
}

func TestGetCommand_InvalidJSON(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`not-json`))
	root.SetArgs([]string{"get", ".x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON input, got nil")
	}
}
