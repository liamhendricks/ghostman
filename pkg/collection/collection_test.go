package collection

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/liamhendricks/ghostman/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdataRoot returns the absolute path to the testdata/.ghostman directory.
func testdataRoot(t *testing.T) string {
	t.Helper()
	// Find the testdata root relative to the test file directory.
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "testdata", ".ghostman")
}

func TestRoots_Default(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	roots, err := Roots(config.Config{})

	require.NoError(t, err)
	assert.Equal(t, []string{
		".ghostman/",
		filepath.Join(home, ".config", "ghostman"),
	}, roots)
}

func TestRoots_WithConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	cfg := config.Config{
		Collections: []string{"~/shared/api/"},
	}
	roots, err := Roots(cfg)

	require.NoError(t, err)
	assert.Equal(t, []string{
		".ghostman/",
		filepath.Join(home, ".config", "ghostman"),
		filepath.Join(home, "shared/api/"),
	}, roots)
}

func TestRoots_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	cfg := config.Config{
		Collections: []string{"~/some/path/"},
	}
	roots, err := Roots(cfg)

	require.NoError(t, err)
	for _, root := range roots {
		assert.False(t, strings.HasPrefix(root, "~/"), "root should not start with ~/: %s", root)
	}
	assert.Contains(t, roots, filepath.Join(home, "some/path/"))
}

func TestFind_ExistingRequest(t *testing.T) {
	root := testdataRoot(t)
	roots := []string{root}

	result, err := Find("auth/login", roots)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "auth", "login.md"), result.FilePath)
	assert.Equal(t, root, result.CollectionRoot)
	assert.Equal(t, filepath.Join(root, "auth"), result.CollectionDir)
}

func TestFind_NotFound(t *testing.T) {
	root := testdataRoot(t)
	roots := []string{root}

	_, err := Find("nonexistent/req", roots)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFind_FirstRootWins(t *testing.T) {
	root := testdataRoot(t)

	// Create a second root in a temp dir with the same auth/login.md
	dir := t.TempDir()
	authDir := filepath.Join(dir, "auth")
	require.NoError(t, os.MkdirAll(authDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(authDir, "login.md"), []byte("# second"), 0o644))

	// First root should win
	roots := []string{root, dir}
	result, err := Find("auth/login", roots)

	require.NoError(t, err)
	assert.Equal(t, root, result.CollectionRoot)
	assert.Equal(t, filepath.Join(root, "auth", "login.md"), result.FilePath)
}

func TestList_AllRequests(t *testing.T) {
	root := testdataRoot(t)
	roots := []string{root}

	paths, err := List(roots)

	require.NoError(t, err)
	assert.True(t, sort.StringsAreSorted(paths), "paths should be sorted")
	assert.Contains(t, paths, "auth/login")
	assert.Contains(t, paths, "auth/signup")
	assert.Contains(t, paths, "users/list")
}

func TestList_ExcludesEnvFiles(t *testing.T) {
	root := testdataRoot(t)
	roots := []string{root}

	paths, err := List(roots)

	require.NoError(t, err)
	for _, p := range paths {
		assert.False(t, strings.HasPrefix(p, "env/"), "should exclude env/ dir: %s", p)
	}
}

func TestList_ExcludesVarsMd(t *testing.T) {
	root := testdataRoot(t)
	roots := []string{root}

	paths, err := List(roots)

	require.NoError(t, err)
	for _, p := range paths {
		assert.False(t, strings.HasSuffix(p, "/vars"), "should exclude vars.md: %s", p)
	}
}

func TestList_NestedDirectories(t *testing.T) {
	root := testdataRoot(t)
	roots := []string{root}

	paths, err := List(roots)

	require.NoError(t, err)
	assert.Contains(t, paths, "resources/posts/get")
	assert.Contains(t, paths, "resources/posts/create")
	assert.True(t, sort.StringsAreSorted(paths), "paths should be sorted")
}

func TestList_Deduplicates(t *testing.T) {
	root := testdataRoot(t)

	// Create a second root with the same auth/login.md
	dir := t.TempDir()
	authDir := filepath.Join(dir, "auth")
	require.NoError(t, os.MkdirAll(authDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(authDir, "login.md"), []byte("# dup"), 0o644))

	roots := []string{root, dir}
	paths, err := List(roots)

	require.NoError(t, err)

	// Count occurrences of auth/login
	count := 0
	for _, p := range paths {
		if p == "auth/login" {
			count++
		}
	}
	assert.Equal(t, 1, count, "auth/login should appear only once")
}

func TestLoadColVars_Found(t *testing.T) {
	root := testdataRoot(t)
	collectionDir := filepath.Join(root, "auth")

	vars, err := LoadColVars(collectionDir)

	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", vars["base_url"])
}

func TestLoadColVars_NotFound(t *testing.T) {
	root := testdataRoot(t)
	collectionDir := filepath.Join(root, "users")

	vars, err := LoadColVars(collectionDir)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{}, vars)
}
