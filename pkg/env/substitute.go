package env

import (
	"fmt"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// Substitute replaces all {{ns:var}} placeholders in s using vars.
// Returns an error if any placeholder has no namespace or is undefined.
// Valid namespaces are "env" and "col". Unknown namespaces are treated as invalid syntax.
func Substitute(s string, vars Vars) (string, error) {
	var missing []string
	var invalidSyntax []string

	result := varPattern.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[2 : len(match)-2] // strip {{ and }}
		parts := strings.SplitN(inner, ":", 2)
		if len(parts) != 2 {
			invalidSyntax = append(invalidSyntax, match)
			return match
		}
		ns, key := parts[0], parts[1]
		switch ns {
		case "env":
			if v, ok := vars.Env[key]; ok {
				return v
			}
		case "col":
			if v, ok := vars.Col[key]; ok {
				return v
			}
		default:
			invalidSyntax = append(invalidSyntax, match)
			return match
		}
		missing = append(missing, ns+":"+key)
		return match
	})

	if len(invalidSyntax) > 0 {
		return "", fmt.Errorf("invalid variable syntax %v: namespace required (use {{env:var}} or {{col:var}})", invalidSyntax)
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("undefined variables: %s", strings.Join(missing, ", "))
	}
	return result, nil
}
