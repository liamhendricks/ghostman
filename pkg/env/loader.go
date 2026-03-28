package env

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// LoadEnvFile parses a YAML env file and returns the variables as a map.
func LoadEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("env file not found: %s", path)
		}
		return nil, fmt.Errorf("env file %s: %w", path, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("env file %s: %w", path, err)
	}

	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}
