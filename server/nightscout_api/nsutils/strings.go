package nsutils

import (
	"regexp"
	"strings"
)

// CloneMap deep-copies a JSON-like map structure.
func CloneMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch typed := v.(type) {
		case map[string]any:
			out[k] = CloneMap(typed)
		case []any:
			copied := make([]any, len(typed))
			for i, item := range typed {
				if nested, ok := item.(map[string]any); ok {
					copied[i] = CloneMap(nested)
				} else {
					copied[i] = item
				}
			}
			out[k] = copied
		default:
			out[k] = typed
		}
	}
	return out
}

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
