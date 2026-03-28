package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChainCmd_FileNotFound(t *testing.T) {
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
	root.SetArgs([]string{"chain", "nonexistent"})

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent chain file, got nil")
	}
	if !strings.Contains(err.Error(), "chain file not found") {
		t.Errorf("error %q does not contain %q", err.Error(), "chain file not found")
	}
}

func TestChainCmd_MissingSteps(t *testing.T) {
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".ghostman", "chain")
	if err := os.MkdirAll(chainDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write a chain file with frontmatter but no steps key
	content := "---\ntitle: test\n---\n"
	if err := os.WriteFile(filepath.Join(chainDir, "empty.md"), []byte(content), 0644); err != nil {
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
	root.SetArgs([]string{"chain", "empty"})

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for missing steps, got nil")
	}
	if !strings.Contains(err.Error(), "missing or empty 'steps'") {
		t.Errorf("error %q does not contain %q", err.Error(), "missing or empty 'steps'")
	}
}

func TestChainCmd_EmptySteps(t *testing.T) {
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".ghostman", "chain")
	if err := os.MkdirAll(chainDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write a chain file with an explicitly empty steps list
	content := "---\nsteps: []\n---\n"
	if err := os.WriteFile(filepath.Join(chainDir, "nosteps.md"), []byte(content), 0644); err != nil {
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
	root.SetArgs([]string{"chain", "nosteps"})

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for empty steps list, got nil")
	}
	if !strings.Contains(err.Error(), "missing or empty 'steps'") {
		t.Errorf("error %q does not contain %q", err.Error(), "missing or empty 'steps'")
	}
}

func TestChainCmd_NoArgs(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"chain"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no chain name provided, got nil")
	}
}
