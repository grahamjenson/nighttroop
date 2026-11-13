package flow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type HTTPAssert interface {
	Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error
}

///////
// Status
///////

type HTTPAssertStatus struct {
	Path string
	Line int
	Code int
}

func (a HTTPAssertStatus) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	outputRaw, _ := MarshalJsonWithoutEscapeHTML(output)

	message := fmt.Sprintf("request %s expected status %d but got %d with body %v", name, a.Code, statusCode, string(outputRaw))
	AssertEqual(t, a.Code, statusCode, message, a.Path, a.Line)

	return nil
}

// jsonHeaders converts headers to plain JSON types so "#" lookups compare like body values.
func jsonHeaders(t *testing.T, headers map[string][]string) map[string]any {
	b, err := json.Marshal(headers)
	if err != nil {
		t.Fatalf("Some weird Marshalerror happened in templateDig: %v", err)
	}
	var headersConverted map[string]any
	if err := json.Unmarshal(b, &headersConverted); err != nil {
		t.Fatalf("Some weird Unmarshal error happened in templateDig: %v", err)
	}
	return headersConverted
}

////////
// Equals and Contains
////////

// AssertEquals is ASSERT_EQUALS <expected> <actual>, e.g. ASSERT_EQUALS 101 "@0.sgv".
type AssertEquals struct {
	Path     string
	Line     int
	Expected any
	Actual   string
}

func (a AssertEquals) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	h := jsonHeaders(t, headers)
	expected := normalizeStringSearch(a.Expected, output, h, ctx)
	actual := normalizeStringSearch(a.Actual, output, h, ctx)
	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(output)
	message := fmt.Sprintf("ASSERT_EQUALS failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, a.Actual, prettyOutput)
	AssertEqual(t, expected, actual, message, a.Path, a.Line)
	return nil
}

// AssertContains is ASSERT_CONTAINS <expected> <actual>, e.g. ASSERT_CONTAINS "not a valid URL" "@message".
type AssertContains struct {
	Path     string
	Line     int
	Expected any
	Actual   string
}

func (a AssertContains) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	h := jsonHeaders(t, headers)
	expected := normalizeStringSearch(a.Expected, output, h, ctx)
	actual := normalizeStringSearch(a.Actual, output, h, ctx)
	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(output)
	message := fmt.Sprintf("ASSERT_CONTAINS failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, a.Actual, prettyOutput)
	AssertContainsValue(t, formatAnyAsString(expected), actual, message, a.Path, a.Line)
	return nil
}

// parseAssertArgs splits `<expected> <actual>`: expected is a quoted string,
// number, true, false or null (for a value that is null or missing), and actual
// is a quoted lookup like "@result.id".
func parseAssertArgs(s string) (any, string, error) {
	s = strings.TrimSpace(s)
	var expected any
	if strings.HasPrefix(s, `"`) {
		quoted, err := strconv.QuotedPrefix(s)
		if err != nil {
			return nil, "", err
		}
		expected, _ = strconv.Unquote(quoted)
		s = s[len(quoted):]
	} else {
		raw, rest, _ := strings.Cut(s, " ")
		s = rest
		if raw == "true" || raw == "false" {
			expected = raw == "true"
		} else if raw == "null" {
			expected = nil
		} else if f, err := strconv.ParseFloat(raw, 64); err == nil {
			expected = f
		} else {
			return nil, "", fmt.Errorf("expected value %q must be quoted, a number, true, false or null", raw)
		}
	}
	actual, err := strconv.Unquote(strings.TrimSpace(s))
	if err != nil {
		return nil, "", fmt.Errorf("actual value %q must be a quoted string", strings.TrimSpace(s))
	}
	return expected, actual, nil
}

////////
// Exists and Length
////////

// AssertExists is ASSERT_EXISTS <actual>, e.g. ASSERT_EXISTS "@0._id".
type AssertExists struct {
	Path   string
	Line   int
	Actual string
}

func (a AssertExists) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	actual := normalizeStringSearch(a.Actual, output, jsonHeaders(t, headers), ctx)
	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(output)
	message := fmt.Sprintf("ASSERT_EXISTS failed: expected value to exist, got nil from [%v], output: %s", a.Actual, prettyOutput)
	AssertNotNil(t, actual, message, a.Path, a.Line)
	return nil
}

// AssertLength is ASSERT_LENGTH <expected> <actual>, e.g. ASSERT_LENGTH 3 "@".
// Strings, arrays and objects have a length; anything else counts as -1.
type AssertLength struct {
	Path     string
	Line     int
	Expected int
	Actual   string
}

func (a AssertLength) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	actual := normalizeStringSearch(a.Actual, output, jsonHeaders(t, headers), ctx)
	actualLength := -1
	if value := reflect.ValueOf(actual); value.IsValid() {
		switch value.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			actualLength = value.Len()
		}
	}
	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(output)
	message := fmt.Sprintf("ASSERT_LENGTH failed: expected length %d, got %d from [%v], output: %s", a.Expected, actualLength, a.Actual, prettyOutput)
	AssertEqual(t, a.Expected, actualLength, message, a.Path, a.Line)
	return nil
}

func MarshalJsonWithoutEscapeHTML(v any) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func normalizeStringSearch(expected, localResult any, headers map[string]any, ctx map[string]any) any {

	// Allow the expected to look through all ctx
	if s, ok := expected.(string); ok {
		if strings.HasPrefix(s, ".") {
			// look up in the context
			expected = Dig(ctx, s[1:])
		} else if strings.HasPrefix(s, "@") {
			if s == "@" {
				return localResult
			}
			// look up in localResult
			expected = Dig(localResult, s[1:])
		} else if strings.HasPrefix(s, "#") {
			// look up in headers
			expected = Dig(headers, s[1:])
		}
	}

	expected = normalizeValue(expected)

	return expected
}

func normalizeValue(v any) any {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)

	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)

	case float32:
		return float64(x)

	case []any:
		for i := range x {
			x[i] = normalizeValue(x[i])
		}
		return x

	case map[string]any:
		for k, v := range x {
			x[k] = normalizeValue(v)
		}
		return x

	default:
		return v
	}
}

func formatAnyAsString(v any) string {
	switch v := v.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func AssertContainsValue(t *testing.T, expected, actual any, message string, file string, line int) bool {
	t.Helper()

	contains := false

	switch a := actual.(type) {
	case string:
		e, ok := expected.(string)
		contains = ok && strings.Contains(a, e)

	case []string:
		e, ok := expected.(string)
		if ok {
			for _, v := range a {
				if v == e {
					contains = true
					break
				}
			}
		}

	case map[string][]string:
		e, ok := expected.(string)
		contains = ok
		if contains {
			_, contains = a[e]
		}

	default:
		// Fall back to equality if you don't want unsupported types
		// to silently pass.
		contains = reflect.DeepEqual(expected, actual)
	}

	if contains {
		return true
	}

	t.Fatal(failure("ASSERT FAILED", file, line, expected, actual, message))

	return false
}

func AssertEqual(t *testing.T, expected, actual any, message string, file string, line int) bool {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return true
	}

	t.Fatal(failure("ASSERT FAILED", file, line, expected, actual, message))

	return false
}

func AssertNotNil(t *testing.T, actual any, message string, file string, line int) bool {
	t.Helper()

	if actual != nil {
		return true
	}

	t.Fatal(failure("ASSERT_EXISTS FAILED", file, line, "a value", actual, message))

	return false
}
