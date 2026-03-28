package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// UpdateColVars creates or merges a key-value pair into the vars.yaml file
// within collectionDir. If the directory or file does not exist, they are created.
func UpdateColVars(collectionDir, key, value string) error {
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		return fmt.Errorf("create collection dir %s: %w", collectionDir, err)
	}

	varsPath := filepath.Join(collectionDir, "vars.yaml")

	var vars map[string]interface{}

	if _, err := os.Stat(varsPath); os.IsNotExist(err) {
		// File does not exist — start with empty map
		vars = map[string]interface{}{}
	} else {
		data, err := os.ReadFile(varsPath)
		if err != nil {
			return fmt.Errorf("vars file %s: %w", varsPath, err)
		}
		if err := yaml.Unmarshal(data, &vars); err != nil || vars == nil {
			vars = map[string]interface{}{}
		}
	}

	vars[key] = value

	out, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("vars file %s: %w", varsPath, err)
	}

	return os.WriteFile(varsPath, out, 0600)
}

// DeriveKey extracts the key name from a gjson path by returning the last
// segment after the final dot. A leading dot is stripped first.
// Examples: ".data.token" -> "token", "status" -> "status", ".status" -> "status"
func DeriveKey(path string) string {
	trimmed := strings.TrimPrefix(path, ".")
	idx := strings.LastIndex(trimmed, ".")
	if idx == -1 {
		return trimmed
	}
	return trimmed[idx+1:]
}
