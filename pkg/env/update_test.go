package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/liamhendricks/ghostman/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEnv_UpdateExistingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "local.yaml")
	err := os.WriteFile(path, []byte("TOKEN: old\nSECRET: keep\n"), 0600)
	require.NoError(t, err)

	err = env.UpdateEnv(path, "TOKEN", "new")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "TOKEN")
	assert.Contains(t, content, "new")
	assert.Contains(t, content, "SECRET")
	assert.Contains(t, content, "keep")
	assert.NotContains(t, content, "old")
}

func TestUpdateEnv_AppendNewKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "local.yaml")
	err := os.WriteFile(path, []byte("A: \"1\"\n"), 0600)
	require.NoError(t, err)

	err = env.UpdateEnv(path, "B", "2")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "A")
	assert.Contains(t, content, "1")
	assert.Contains(t, content, "B")
	assert.Contains(t, content, "2")
}

func TestUpdateEnv_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")

	err := env.UpdateEnv(path, "KEY", "val")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "env file not found")
}

func TestUpdateEnv_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "local.yaml")
	err := os.WriteFile(path, []byte(""), 0600)
	require.NoError(t, err)

	err = env.UpdateEnv(path, "KEY", "val")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "KEY")
	assert.Contains(t, string(data), "val")
}

func TestUpdateEnv_ValueWithSpaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "local.yaml")
	err := os.WriteFile(path, []byte("A: \"1\"\n"), 0600)
	require.NoError(t, err)

	err = env.UpdateEnv(path, "KEY", "hello world")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "KEY")
	assert.Contains(t, string(data), "hello world")
}
