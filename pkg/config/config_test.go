package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, ".ghostmanrc")

	cfg, err := LoadFrom(nonExistentPath)

	require.NoError(t, err)
	assert.Equal(t, Config{}, cfg)
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ghostmanrc")
	content := "collections:\n  - \".ghostman/\"\n  - \"~/shared/api/\"\ndefault_env: staging\n"
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := LoadFrom(path)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Collections: []string{".ghostman/", "~/shared/api/"},
		DefaultEnv:  "staging",
	}, cfg)
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ghostmanrc")
	err := os.WriteFile(path, []byte(""), 0o644)
	require.NoError(t, err)

	cfg, err := LoadFrom(path)

	require.NoError(t, err)
	assert.Equal(t, Config{}, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ghostmanrc")
	content := "collections: [\ndefault_env: [broken yaml"
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	_, err = LoadFrom(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config.yaml")
}

func TestLoad_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ghostmanrc")
	content := "default_env: production\n"
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := LoadFrom(path)

	require.NoError(t, err)
	assert.Equal(t, Config{
		DefaultEnv:  "production",
		Collections: nil,
	}, cfg)
}
