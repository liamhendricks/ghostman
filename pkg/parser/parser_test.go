package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseValidRequest(t *testing.T) {
	spec, err := Parse("../../testdata/valid-request.md")
	require.NoError(t, err)

	require.Equal(t, 1, spec.GhostmanVersion)
	require.Equal(t, "POST", spec.Method)
	require.Equal(t, "{{env:base_url}}", spec.BaseURL)
	require.Equal(t, "/users/:id", spec.Path)

	require.Equal(t, "application/json", spec.Headers["Content-Type"])
	require.Equal(t, "Bearer {{env:token}}", spec.Headers["Authorization"])

	require.Equal(t, "1", spec.QueryParams["page"])
	require.Equal(t, "{{col:filter}}", spec.QueryParams["filter"])

	require.NotNil(t, spec.Body)
	require.Contains(t, string(spec.Body), `"name": "{{env:name}}"`)
	require.Equal(t, "json", spec.BodyLang)
}

func TestParseMissingVersion(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/missing-version.md"
	content := "---\nmethod: GET\nbase_url: http://example.com\npath: /test\n---\n\n# Missing Version\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	_, err := Parse(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ghostman_version")
}

func TestParseUnsupportedVersion(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/unsupported-version.md"
	content := "---\nghostman_version: 2\nmethod: GET\nbase_url: http://example.com\npath: /test\n---\n\n# Unsupported Version\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	_, err := Parse(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported ghostman_version 2")
}

func TestBodyExtraction(t *testing.T) {
	spec, err := Parse("../../testdata/valid-request.md")
	require.NoError(t, err)

	require.NotNil(t, spec.Body)
	require.Equal(t, "json", spec.BodyLang)
}

func TestAmbiguousBody(t *testing.T) {
	_, err := Parse("../../testdata/ambiguous-body.md")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous body")
	require.Contains(t, err.Error(), "2")
	require.Contains(t, err.Error(), "json")
}

func TestBodyEscapeHatch(t *testing.T) {
	spec, err := Parse("../../testdata/body-escape-hatch.md")
	require.NoError(t, err)

	require.NotNil(t, spec.Body)
	require.Contains(t, string(spec.Body), "actual")
	require.Equal(t, "body", spec.BodyLang)
}

func TestNoBody(t *testing.T) {
	spec, err := Parse("../../testdata/no-body-get.md")
	require.NoError(t, err)

	require.Nil(t, spec.Body)
	require.Empty(t, spec.BodyLang)
}

func TestParseCertAndKey(t *testing.T) {
	spec, err := Parse("../../testdata/mtls-request.md")
	require.NoError(t, err)

	require.Equal(t, "{{env:client_cert}}", spec.Cert)
	require.Equal(t, "{{env:client_key}}", spec.Key)
}

func TestParseNoCert(t *testing.T) {
	spec, err := Parse("../../testdata/valid-request.md")
	require.NoError(t, err)

	require.Empty(t, spec.Cert)
	require.Empty(t, spec.Key)
}

func TestNoFrontmatter(t *testing.T) {
	_, err := Parse("../../testdata/no-frontmatter.md")
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "frontmatter") || strings.Contains(err.Error(), "no frontmatter"),
		"expected error to contain 'frontmatter', got: %s", err.Error(),
	)
}
