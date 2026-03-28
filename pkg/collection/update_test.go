package collection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateColVars_CreatesFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	collectionDir := filepath.Join(dir, "auth")

	err := UpdateColVars(collectionDir, "token", "abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	varsPath := filepath.Join(collectionDir, "vars.yaml")
	if _, err := os.Stat(varsPath); os.IsNotExist(err) {
		t.Fatal("expected vars.yaml to be created, but it does not exist")
	}

	content, err := os.ReadFile(varsPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "token") {
		t.Errorf("expected 'token' key in vars.yaml, got: %q", s)
	}
	if !strings.Contains(s, "abc123") {
		t.Errorf("expected 'abc123' value in vars.yaml, got: %q", s)
	}
}

func TestUpdateColVars_MergesWithExistingFile(t *testing.T) {
	dir := t.TempDir()
	collectionDir := filepath.Join(dir, "auth")
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatal(err)
	}
	varsPath := filepath.Join(collectionDir, "vars.yaml")
	if err := os.WriteFile(varsPath, []byte("base_url: http://x\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := UpdateColVars(collectionDir, "token", "new")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	content, err := os.ReadFile(varsPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "base_url") {
		t.Errorf("expected 'base_url' key preserved, got: %q", s)
	}
	if !strings.Contains(s, "http://x") {
		t.Errorf("expected 'http://x' value preserved, got: %q", s)
	}
	if !strings.Contains(s, "token") {
		t.Errorf("expected 'token' key added, got: %q", s)
	}
	if !strings.Contains(s, "new") {
		t.Errorf("expected 'new' value added, got: %q", s)
	}
}

func TestUpdateColVars_OverwritesExistingKey(t *testing.T) {
	dir := t.TempDir()
	collectionDir := filepath.Join(dir, "auth")
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatal(err)
	}
	varsPath := filepath.Join(collectionDir, "vars.yaml")
	if err := os.WriteFile(varsPath, []byte("token: old\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := UpdateColVars(collectionDir, "token", "updated")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	content, err := os.ReadFile(varsPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "updated") {
		t.Errorf("expected 'updated' value, got: %q", s)
	}
	if strings.Contains(s, "old") {
		t.Errorf("expected 'old' value to be overwritten, got: %q", s)
	}
}

func TestUpdateColVars_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	// Use a deeply nested dir that does not exist
	collectionDir := filepath.Join(dir, "nested", "auth")

	err := UpdateColVars(collectionDir, "token", "abc")
	if err != nil {
		t.Fatalf("expected no error when dir does not exist, got: %v", err)
	}

	if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
		t.Fatal("expected collectionDir to be created by UpdateColVars")
	}
}

func TestDeriveKey_DotDataToken(t *testing.T) {
	got := DeriveKey(".data.token")
	if got != "token" {
		t.Errorf("DeriveKey(%q) = %q, want %q", ".data.token", got, "token")
	}
}

func TestDeriveKey_DotItems0Id(t *testing.T) {
	got := DeriveKey(".items.0.id")
	if got != "id" {
		t.Errorf("DeriveKey(%q) = %q, want %q", ".items.0.id", got, "id")
	}
}

func TestDeriveKey_NoDots(t *testing.T) {
	got := DeriveKey("status")
	if got != "status" {
		t.Errorf("DeriveKey(%q) = %q, want %q", "status", got, "status")
	}
}

func TestDeriveKey_LeadingDotOnly(t *testing.T) {
	got := DeriveKey(".status")
	if got != "status" {
		t.Errorf("DeriveKey(%q) = %q, want %q", ".status", got, "status")
	}
}
