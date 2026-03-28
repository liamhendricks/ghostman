package env

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// LoadEnvFile tests

func TestLoadEnvFile(t *testing.T) {
	vars, err := LoadEnvFile("../../testdata/env/staging.yaml")
	require.NoError(t, err)
	require.Equal(t, "https://staging.api.example.com", vars["base_url"])
	require.Equal(t, "staging-token-123", vars["token"])
	require.Equal(t, "test-user", vars["name"])
	require.Equal(t, "v1", vars["api_version"])
}

func TestLoadEnvFileMissing(t *testing.T) {
	_, err := LoadEnvFile("../../testdata/nonexistent.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestLoadColVars(t *testing.T) {
	vars, err := LoadEnvFile("../../testdata/vars.yaml")
	require.NoError(t, err)
	require.Equal(t, "active", vars["filter"])
	require.Equal(t, "42", vars["user_id"])
}

// Load test

func TestLoad(t *testing.T) {
	e, err := Load("../../testdata/env/staging.yaml")
	require.NoError(t, err)
	require.Equal(t, "https://staging.api.example.com", e.EnvVars["base_url"])
	require.Equal(t, "staging-token-123", e.EnvVars["token"])
}

// Substitute tests

func TestSubstituteEnv(t *testing.T) {
	result, err := Substitute("https://{{env:base_url}}/api", Vars{Env: map[string]string{"base_url": "example.com"}})
	require.NoError(t, err)
	require.Equal(t, "https://example.com/api", result)
}

func TestSubstituteCol(t *testing.T) {
	result, err := Substitute("{{col:filter}}", Vars{Col: map[string]string{"filter": "active"}})
	require.NoError(t, err)
	require.Equal(t, "active", result)
}

func TestSubstituteMultiple(t *testing.T) {
	result, err := Substitute("{{env:a}} and {{col:b}}", Vars{
		Env: map[string]string{"a": "1"},
		Col: map[string]string{"b": "2"},
	})
	require.NoError(t, err)
	require.Equal(t, "1 and 2", result)
}

func TestSubstituteNoNamespace(t *testing.T) {
	_, err := Substitute("{{var}}", Vars{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "namespace required")
}

func TestSubstituteMissing(t *testing.T) {
	_, err := Substitute("{{env:missing}}", Vars{Env: map[string]string{}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined variables")
	require.Contains(t, err.Error(), "env:missing")
}

func TestSubstituteMultipleMissing(t *testing.T) {
	_, err := Substitute("{{env:a}} {{col:b}}", Vars{Env: map[string]string{}, Col: map[string]string{}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "env:a")
	require.Contains(t, err.Error(), "col:b")
}

func TestSubstituteUnknownNamespace(t *testing.T) {
	_, err := Substitute("{{req:var}}", Vars{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "namespace")
}
