package nsutils

import (
	"regexp"
	"strings"
)

// PathValue returns a nested value from a JSON-like object using dot or bracket
// notation.
func PathValue(data map[string]any, path string) any {
	if data == nil || path == "" {
		return nil
	}
	if value, ok := data[path]; ok {
		return value
	}
	parts := splitPath(path)
	var current any = data
	for _, part := range parts {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current, ok = object[part]
		if !ok {
			return nil
		}
	}
	return current
}

func splitPath(path string) []string {
	if strings.Contains(path, ".") {
		return strings.Split(path, ".")
	}
	brackets := regexp.MustCompile(`\[([^\]]+)\]`)
	if brackets.MatchString(path) {
		path = brackets.ReplaceAllString(path, `.$1`)
		path = strings.TrimPrefix(path, ".")
		return strings.Split(path, ".")
	}
	return []string{path}
}

// ValidateIdentifierIsntBad checks if the identifier is valid and returns an error if it is not.
// Supports up to UUID 36 chars, alpha numeric and _ and - only
// This is because users are allowed to set their own identifiers, but we have limits
func ValidateIdentifierIsntBad(identifier string) *HTTPError {
	if len(identifier) == 0 {
		return BadIdentifierError
	}

	if len(identifier) > 36 {
		return BadIdentifierError
	}

	// Below code is basically this, but just faster
	// if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(identifier) {
	// 	return BadIdentifierError
	// }

	for i := 0; i < len(identifier); i++ {
		c := identifier[i]
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' ||
			c == '-' {
			continue
		}
		return BadIdentifierError
	}
	return nil
}
