package bdd_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/data/db"
)

type Action interface {
	ToPathLine() string
	ToFlatString() string
	PrintAction()
	Execute(t *testing.T, page playwright.Page, ctx map[string]any) error
	BuildAndValidate(t *testing.T, ctx map[string]any) error
}

type ActionFile struct {
	Name    string
	Actions []Action
}

//////////////////
// ACTIONS
//////////////////

// /////
// GOTO
// //////
type Goto struct {
	URL string

	AssertStatus *AssertStatus
	Asserts      []*Assert

	Path string
	Line int
}

func (r *Goto) ToFlatString() string {
	return fmt.Sprintf("GOTO %s", r.URL)
}

func (r *Goto) ToPathLine() string {
	return fmt.Sprintf("%s:%d", r.Path, r.Line)
}

func (r *Goto) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Goto) Execute(t *testing.T, page playwright.Page, ctx map[string]any) error {
	// Implement the execution logic for the Goto action here
	return nil
}

func (r *Goto) BuildAndValidate(t *testing.T, ctx map[string]any) error {
	url, err := renderTemplate(r.URL, ctx)
	if err != nil {
		return err
	}
	r.URL = url

	return nil
}

//////
// CLICK
//////

type Click struct {
	Locator string

	AssertStatus *AssertStatus
	Asserts      []*Assert

	Path string
	Line int
}

func (r *Click) ToFlatString() string {
	return fmt.Sprintf("CLICK %s", r.Locator)
}

func (r *Click) ToPathLine() string {
	return fmt.Sprintf("%s:%d", r.Path, r.Line)
}

func (r *Click) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Click) Execute(t *testing.T, page playwright.Page, ctx map[string]any) error {
	// Implement the execution logic for the Click action here
	return nil
}

func (r *Click) BuildAndValidate(t *testing.T, ctx map[string]any) error {
	// Implement the logic to build the real action based on the context here
	return nil
}

/////
// INPUT
//////

type Input struct {
	Locator string
	Value   string

	Path string
	Line int
}

func (r *Input) ToFlatString() string {
	return fmt.Sprintf("INPUT %s %s", r.Locator, r.Value)
}

func (r *Input) ToPathLine() string {
	return fmt.Sprintf("%s:%d", r.Path, r.Line)
}

func (r *Input) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Input) Execute(t *testing.T, page playwright.Page, ctx map[string]any) error {
	// Implement the execution logic for the Input action here
	return nil
}

func (r *Input) BuildAndValidate(t *testing.T, ctx map[string]any) error {
	// Implement the logic to build the real action based on the context here
	return nil
}

// ///
// SET
// ///
type Set struct {
	Key   string
	Value string

	Path string
	Line int
}

func (r *Set) ToFlatString() string {
	return fmt.Sprintf("SET %s %s", r.Key, r.Value)
}

func (r *Set) ToPathLine() string {
	return fmt.Sprintf("%s:%d", r.Path, r.Line)
}

func (r *Set) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Set) Execute(t *testing.T, page playwright.Page, ctx map[string]any) error {
	// Implement the execution logic for the Set action here
	return nil
}

func (r *Set) BuildAndValidate(t *testing.T, ctx map[string]any) error {
	// Implement the logic to build the real action based on the context here
	return nil
}

// ///
// Non Actions
// ///

type Assert struct {
	Expression string

	Path string
	Line int
}

func (r *Assert) Execute(t *testing.T, ctx map[string]any) error {

	return nil
}

type AssertStatus struct {
	Code int

	Path string
	Line int
}

func ExecuteAsserts(t *testing.T, assertStructs []*Assert, assertStatus *AssertStatus, realStatusCode int, output string, ctx map[string]any) {
	// Test the asserts against the real value returns like page.Visible, and response.Headers...
}

// //////
//
// //////
func ParseActionFile(fileContents string, ctx map[string]any, path string) (*ActionFile, error) {
	actions, err := parseFile(strings.NewReader(fileContents), path)
	if err != nil {
		return nil, err
	}

	// name is the file name without folders or extension
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return &ActionFile{
		Name:    name,
		Actions: actions,
	}, nil
}

func parseFile(r io.Reader, path string) ([]Action, error) {
	scanner := bufio.NewScanner(r)

	var actions []Action
	var current Action

	fmt.Println("Parsing action file:", path)
	flush := func() {
		if current != nil {
			actions = append(actions, current)
			current = nil
		}
	}
	lineNumber := 0
	for scanner.Scan() {
		lineNumber += 1
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		split := strings.Split(line, " ")
		if len(split) == 0 {
			continue
		}

		switch split[0] {
		case "GOTO":
			flush()
			if len(split) != 2 {
				return nil, fmt.Errorf("invalid goto action line: %q", line)
			}
			url := split[1]
			if url == "" {
				return nil, fmt.Errorf("empty goto url")
			}

			current = &Goto{
				Path:    path,
				Line:    lineNumber,
				URL:     url,
				Asserts: []*Assert{},
			}
		case "CLICK":
			flush()
			if len(split) != 2 {
				return nil, fmt.Errorf("invalid click action line: %q", line)
			}
			locator := split[1]
			if locator == "" {
				return nil, fmt.Errorf("empty click locator")
			}

			current = &Click{
				Path:    path,
				Line:    lineNumber,
				Locator: locator,
				Asserts: []*Assert{},
			}

		case "SET":
			flush()
			if len(split) != 3 {
				return nil, fmt.Errorf("invalid set action line: %q", line)
			}
			key := split[1]
			value := split[2]

			if value == "" {
				return nil, fmt.Errorf("empty set value")
			}

			current = &Set{
				Path:  path,
				Line:  lineNumber,
				Key:   key,
				Value: value,
			}
		// Addint to action
		case "ASSERT_STATUS":
			if current == nil {
				return nil, fmt.Errorf("assert_status before action")
			}

			if len(split) != 2 {
				return nil, fmt.Errorf("invalid assert_status action line: %q", line)
			}

			statusCode, err := strconv.Atoi(split[1])
			if err != nil {
				return nil, fmt.Errorf("invalid assert_status: %q", split[1])
			}
			assertStatus := &AssertStatus{
				Path: path,
				Line: lineNumber,
				Code: statusCode,
			}
			switch v := current.(type) {
			case *Goto:
				// valid types for assert_status
				v.AssertStatus = assertStatus
			case *Click:
				// valid types for assert_status
				v.AssertStatus = assertStatus
			default:
				return nil, fmt.Errorf("assert_status not applicable for this action")
			}
		case "ASSERT":
			if current == nil {
				return nil, fmt.Errorf("assert before action")
			}

			if len(split) != 2 {
				return nil, fmt.Errorf("invalid assert action line: %q", line)
			}

			assertStatus := &Assert{
				Expression: split[1],
				Path:       path,
				Line:       lineNumber,
			}
			switch v := current.(type) {
			case *Goto:
				// valid types for assert
				v.Asserts = append(v.Asserts, assertStatus)
			case *Click:
				// valid types for assert
				v.Asserts = append(v.Asserts, assertStatus)
			default:
				return nil, fmt.Errorf("assert not applicable for this action")
			}

		default:
			return nil, fmt.Errorf("unknown line: %q", line)
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return actions, nil
}

func templateTime(offset int) string {
	// top of the minute
	t := time.Now().Add(-time.Duration(offset) * time.Minute)
	return t.Format(time.RFC3339)
}

func templateITime(offset int) int64 {
	// top of the minute
	t := time.Now().Add(-time.Duration(offset) * time.Minute)
	return t.UnixMilli()
}

func templateDig(root any, path string) any {

	parts := strings.Split(path, ".")
	var cur = root

	for _, p := range parts {
		switch v := cur.(type) {

		case map[string]any:
			next, ok := v[p]
			if !ok {
				return nil
			}
			cur = next

		case []any:
			// allow numeric indexing into slices
			i, err := strconv.Atoi(p)
			if err != nil || i < 0 || i >= len(v) {
				return nil
			}
			cur = v[i]

		default:
			return nil
		}
	}

	return cur
}

func renderTemplate(input string, ctx map[string]any) (string, error) {
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				dug := templateDig(ctx, path)
				switch dug.(type) {
				case nil:
					return "no_value"
				case string:
					return dug.(string)
				case float64:
					return fmt.Sprintf("%.0f", dug.(float64))
				case int:
					return dug.(int)
				case bool:
					return dug.(bool)
				default:
					return dug
				}
			},
			"site": func() string {
				if v, ok := ctx["site"]; ok && v != nil {
					return ctx["site"].(string)
				}
				return ctx["default_site"].(string)
			},
			"token": func() string {
				if v, ok := ctx["token"]; ok && v != nil {
					return ctx["token"].(string)
				}
				return ctx["default_token"].(string)
			},
			"read_token": func() string {
				return ctx["read_token"].(string)
			},
			"jwt": func() string {
				if v, ok := ctx["jwt"]; ok && v != nil {
					return ctx["jwt"].(string)
				}
				return ctx["default_jwt"].(string)
			},
			"read_jwt": func() string {
				return ctx["read_jwt"].(string)
			},
			"time":  templateTime,
			"itime": templateITime,
			"tokenByName": func(list []any, name string) string {
				for _, item := range list {
					m := item.(map[string]any)
					if m["name"] == name {
						return m["accessToken"].(string)
					}
				}
				return ""
			},
		}).
		Parse(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func assertTemplateCheck(assertStruct Assert, ctx map[string]any, localResult any, headers map[string][]string, t *testing.T) (string, error) {
	// Need to forse the headers to be json-esk so types work
	b, err := json.Marshal(headers)
	if err != nil {
		t.Fatalf("Some weird Marshalerror happened in templateDig: %v", err)
		return "false", err
	}
	var headersConverted map[string]any
	err = json.Unmarshal(b, &headersConverted)
	if err != nil {
		t.Fatalf("Some weird Unmarshal error happened in templateDig: %v", err)
		return "false", err
	}

	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(localResult)
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				return templateDig(ctx, path)
			},
			// TESTING FUNCTIONS
			"assertEquals": func(rawExpected, rawActual any) string {
				expected := normalizeStringSearch(rawExpected, localResult, headersConverted, ctx)
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				message := fmt.Sprintf("assertEquals failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, rawActual, prettyOutput)
				AssertEqual(t, expected, actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
			"assertContains": func(rawExpected, rawActual any) string {
				expected := normalizeStringSearch(rawExpected, localResult, headersConverted, ctx)
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				message := fmt.Sprintf("assertContains failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, rawActual, prettyOutput)
				AssertContains(t, formatAnyAsString(expected), actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
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
		Parse(assertStruct.Expression)
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

func AssertContains(t *testing.T, expected, actual any, message string, file string, line int) bool {
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

func RunTheActionFileAsATest(t *testing.T, root string, path string, context playwright.BrowserContext, siteInformation *db.SiteCreateResult, defaultJwt string) {
	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}

	removeNightscout := strings.TrimPrefix(path, root+"/")
	t.Run(removeNightscout, func(t *testing.T) {
		actionFileBytes, err := embeddedFiles.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read action file: %v", err)
		}
		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running action file: %s\n", path)
		fmt.Println("===============================================")

		actionContext := map[string]any{
			"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
			"default_token": siteInformation.PwdNSToken,
			"default_jwt":   defaultJwt,
		}
		actionFile, err := ParseActionFile(string(actionFileBytes), actionContext, path)
		if err != nil {
			t.Fatalf("failed to parse action file: %v", err)
		}
		for _, action := range actionFile.Actions {
			err := action.BuildAndValidate(t, actionContext)
			if err != nil {
				t.Fatalf("failed to build real action: %v", err)
			}

			fmt.Println()
			fmt.Println("=============== " + action.ToFlatString() + "  ================")
			fmt.Println()
			action.PrintAction()
			fmt.Println()
			fmt.Println("- - - - - - - - - - - - - - - - ")
			fmt.Println()
			// default information

			err = action.Execute(t, page, actionContext)
			if err != nil {
				t.Fatalf("failed to execute action: %v, action: %s", err, action.ToFlatString())
			}

			fmt.Println("===============================")
		}

	})
}
