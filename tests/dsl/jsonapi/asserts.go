package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"
)

type Assert interface {
	Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error
}

///////
// Status
///////

type AssertStatus struct {
	Path string
	Line int
	Code int
}

func (a AssertStatus) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	outputRaw, _ := MarshalJsonWithoutEscapeHTML(output)

	message := fmt.Sprintf("request %s expected status %d but got %d with body %v", name, a.Code, statusCode, string(outputRaw))
	AssertEqual(t, a.Code, statusCode, message, a.Path, a.Line)

	return nil
}

////////
// Template
////////

type AssertTemplate struct {
	Path string
	Line int
	Raw  string
}

func (a AssertTemplate) Execute(t *testing.T, name string, statusCode int, output any, headers http.Header, ctx map[string]any) error {
	outputRaw, _ := MarshalJsonWithoutEscapeHTML(output)

	_, err := assertTemplateCheck(a, ctx, output, headers, t)
	if err != nil {
		t.Fatalf("assert Template Error %s assert %s output %v render failed: %v", name, a.Raw, string(outputRaw), err)
	}
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
// number, true or false, and actual is a quoted lookup like "@result.id".
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
		} else if f, err := strconv.ParseFloat(raw, 64); err == nil {
			expected = f
		} else {
			return nil, "", fmt.Errorf("expected value %q must be quoted, a number, true or false", raw)
		}
	}
	actual, err := strconv.Unquote(strings.TrimSpace(s))
	if err != nil {
		return nil, "", fmt.Errorf("actual value %q must be a quoted string", strings.TrimSpace(s))
	}
	return expected, actual, nil
}

func assertTemplateCheck(assertStruct AssertTemplate, ctx map[string]any, localResult any, headers map[string][]string, t *testing.T) (string, error) {
	headersConverted := jsonHeaders(t, headers)

	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(localResult)
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				return templateDig(ctx, path)
			},
			// TESTING FUNCTIONS
			"assertExists": func(rawActual any) string {
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				message := fmt.Sprintf("assertExists failed: expected value to exist, got nil from [%v], output: %s", rawActual, prettyOutput)
				AssertNotNil(t, actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
			"assertLength": func(expectedLength int, rawActual any) string {
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)

				actualLength := -1

				value := reflect.ValueOf(actual)

				if value.IsValid() {
					switch value.Kind() {
					case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
						actualLength = value.Len()
					}
				}

				message := fmt.Sprintf(
					"assertLength failed: expected length %d, got %d from [%v], output: %s",
					expectedLength,
					actualLength,
					rawActual,
					prettyOutput,
				)

				AssertEqual(
					t,
					expectedLength,
					actualLength,
					message,
					assertStruct.Path,
					assertStruct.Line,
				)

				return "true"
			},
		}).
		Parse(assertStruct.Raw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
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
			expected = templateDig(ctx, s[1:])
		} else if strings.HasPrefix(s, "@") {
			if s == "@" {
				return localResult
			}
			// look up in localResult
			expected = templateDig(localResult, s[1:])
		} else if strings.HasPrefix(s, "#") {
			// look up in headers
			expected = templateDig(headers, s[1:])
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

	t.Fatalf(
		"\n ASSERT FAILED \n %s:%d: %s\nexpected to contain: %#v\nactual:              %#v",
		file,
		line,
		message,
		expected,
		actual,
	)

	return false
}

func AssertEqual(t *testing.T, expected, actual any, message string, file string, line int) bool {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return true
	}

	t.Fatalf(
		"\n ASSERT EQUALS FAILED \n %s:%d: %s\nexpected: %#v\nactual:   %#v",
		file,
		line,
		message,
		expected,
		actual,
	)

	return false
}

func AssertNotNil(t *testing.T, actual any, message string, file string, line int) bool {
	t.Helper()

	if actual != nil {
		return true
	}

	t.Fatalf(
		"\n ASSERT NOT NIL FAILED \n %s:%d: %s\nactual:   %#v",
		file,
		line,
		message,
		actual,
	)

	return false
}
