package env

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// UpdateEnv reads the YAML env file at path, sets key=value (updating if key exists,
// appending if not), and writes the file back. The file must already exist.
func UpdateEnv(path, key, value string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("env file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("env file %s: %w", path, err)
	}

	var vars map[string]interface{}
	if err := yaml.Unmarshal(data, &vars); err != nil || vars == nil {
		vars = map[string]interface{}{}
	}

	vars[key] = value

	out, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("env file %s: %w", path, err)
	}

	return os.WriteFile(path, out, 0600)
}
