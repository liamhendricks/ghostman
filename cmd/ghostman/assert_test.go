package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestAssertCommand_MatchPassesThrough(t *testing.T) {
	jsonInput := `{"status":"ok"}`
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(jsonInput))
	root.SetArgs([]string{"assert", `.status == "ok"`})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, `"status":"ok"`) {
		t.Errorf("expected original JSON on stdout (pass-through), got: %q", out)
	}
}

func TestAssertCommand_MismatchExitsNonZero(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"status":"ok"}`))
	root.SetArgs([]string{"assert", `.status == "fail"`})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit on assertion failure, got nil")
	}

	if !strings.Contains(err.Error(), "assertion failed") {
		t.Errorf("expected 'assertion failed' in error, got: %v", err)
	}
}

func TestAssertCommand_NotEqualPassesThrough(t *testing.T) {
	jsonInput := `{"status":"ok"}`
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(jsonInput))
	root.SetArgs([]string{"assert", `.status != "fail"`})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected success for != assertion, got error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, `"status":"ok"`) {
		t.Errorf("expected original JSON on stdout (pass-through), got: %q", out)
	}
}

func TestAssertCommand_NotEqualMismatchExitsNonZero(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`{"status":"ok"}`))
	root.SetArgs([]string{"assert", `.status != "ok"`})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit when != condition is false, got nil")
	}
}

func TestAssertCommand_InvalidJSON(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(bytes.NewBufferString(`not-json`))
	root.SetArgs([]string{"assert", `.x == "y"`})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON input, got nil")
	}
}
