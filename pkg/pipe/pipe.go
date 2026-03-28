package pipe

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/tidwall/gjson"
)

// ReadJSON reads all bytes from r and validates it is JSON.
// Returns the raw bytes on success.
func ReadJSON(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no input: expected JSON on stdin")
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON: input is not valid JSON")
	}
	return data, nil
}

// Extract returns the value at the given gjson path from JSON data.
// Returns an error if the path does not exist in the JSON.
// The path may optionally have a leading "." which will be stripped.
func Extract(data []byte, path string) (string, error) {
	trimmed := strings.TrimPrefix(path, ".")
	result := gjson.GetBytes(data, trimmed)
	if !result.Exists() {
		return "", fmt.Errorf("path not found: %s", path)
	}
	return result.String(), nil
}

// Assert evaluates an assertion expression against JSON data.
// Supported operators: == and !=
// Expression format: .path == "value" or .path != "value"
// Returns nil on success, error with diagnostic on failure.
func Assert(data []byte, expr string) error {
	var path, expected, op string

	if idx := strings.Index(expr, " == "); idx != -1 {
		path = strings.TrimSpace(expr[:idx])
		expected = strings.TrimSpace(expr[idx+4:])
		op = "=="
	} else if idx := strings.Index(expr, " != "); idx != -1 {
		path = strings.TrimSpace(expr[:idx])
		expected = strings.TrimSpace(expr[idx+4:])
		op = "!="
	} else {
		// Check for unsupported operators first
		for _, unsupported := range []string{" > ", " < ", " >= ", " <= "} {
			if strings.Contains(expr, unsupported) {
				return fmt.Errorf("unsupported operator in expression: %q (supported: ==, !=)", expr)
			}
		}
		return fmt.Errorf("invalid assertion expression: %q (expected format: .path == \"value\" or .path != \"value\")", expr)
	}

	// Strip surrounding quotes from expected value if present
	if len(expected) >= 2 && expected[0] == '"' && expected[len(expected)-1] == '"' {
		expected = expected[1 : len(expected)-1]
	}

	actual, err := Extract(data, path)
	if err != nil {
		return fmt.Errorf("assertion failed: %w", err)
	}

	switch op {
	case "==":
		if actual != expected {
			return fmt.Errorf("assertion failed: %s == %q (got: %q)", path, expected, actual)
		}
	case "!=":
		if actual == expected {
			return fmt.Errorf("assertion failed: %s != %q (got: %q)", path, expected, actual)
		}
	}

	return nil
}
