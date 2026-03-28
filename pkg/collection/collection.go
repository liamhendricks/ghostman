package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/liamhendricks/ghostman/pkg/config"
	"github.com/liamhendricks/ghostman/pkg/env"
)

// FindResult holds the result of locating a request file in a collection root.
type FindResult struct {
	FilePath       string // absolute path to the .md file
	CollectionRoot string // the root dir that contained this file
	CollectionDir  string // the subdirectory within root (e.g. full path to "auth" dir)
}

// expandPath expands a leading ~/ to the user's home directory.
func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

// Roots returns the ordered list of collection roots.
// It always includes ".ghostman/" (cwd-relative) and ~/.ghostman/ as defaults,
// followed by any paths in cfg.Collections (with ~/ expanded).
func Roots(cfg config.Config) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	roots := []string{
		".ghostman/",
		filepath.Join(home, ".config", "ghostman"),
	}

	for _, col := range cfg.Collections {
		expanded, err := expandPath(col)
		if err != nil {
			return nil, err
		}
		roots = append(roots, expanded)
	}

	return roots, nil
}

// Find locates the first matching .md file for collectionName (e.g. "auth/login")
// across the provided roots. Returns FindResult with the absolute file path,
// collection root, and collection subdirectory.
func Find(collectionName string, roots []string) (FindResult, error) {
	parts := strings.SplitN(collectionName, "/", 2)
	if len(parts) != 2 {
		return FindResult{}, fmt.Errorf("request %q not found in any collection", collectionName)
	}
	dir, name := parts[0], parts[1]

	for _, root := range roots {
		candidate := filepath.Join(root, dir, name+".md")
		if _, err := os.Stat(candidate); err == nil {
			absPath, err := filepath.Abs(candidate)
			if err != nil {
				return FindResult{}, fmt.Errorf("abs path: %w", err)
			}
			absRoot, err := filepath.Abs(root)
			if err != nil {
				return FindResult{}, fmt.Errorf("abs root: %w", err)
			}
			return FindResult{
				FilePath:       absPath,
				CollectionRoot: absRoot,
				CollectionDir:  filepath.Join(absRoot, dir),
			}, nil
		}
	}

	return FindResult{}, fmt.Errorf("request %q not found in any collection", collectionName)
}

// List returns a sorted, deduplicated list of collection/name paths from all roots.
// It excludes files in "env/" subdirectories and files named "vars.md".
// Paths are relative to the root (e.g. "auth/login", "jsonplaceholder/posts/get").
func List(roots []string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string

	for _, root := range roots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			// Skip the env/ top-level directory entirely
			if d.IsDir() && rel == "env" {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			nameWithoutExt := strings.TrimSuffix(rel, ".md")
			if filepath.Base(nameWithoutExt) == "vars" {
				return nil
			}
			// Normalise to forward slashes on all platforms
			normalized := filepath.ToSlash(nameWithoutExt)
			if !seen[normalized] {
				seen[normalized] = true
				result = append(result, normalized)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking root %s: %w", root, err)
		}
	}

	sort.Strings(result)
	return result, nil
}

// LoadColVars loads variables from a vars.md file in the given collection directory.
// If no vars.md exists, an empty map and nil error are returned.
func LoadColVars(collectionDir string) (map[string]string, error) {
	varsPath := filepath.Join(collectionDir, "vars.yaml")
	if _, err := os.Stat(varsPath); os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	return env.LoadEnvFile(varsPath)
}
