package bdd_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/data/db"
)

type Action interface {
	ToErrorLine(msg string) string
	ToFlatString() string
	PrintAction()
	Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error)
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

	Asserts []Assert

	Path string
	Line int
}

func (r *Goto) ToFlatString() string {
	return fmt.Sprintf("GOTO %s", r.URL)
}

func (r *Goto) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.Path, r.Line, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.Path, r.Line, msg, r.ToFlatString())
}

func (r *Goto) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Goto) AddAssert(assert Assert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Goto) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {

	url, err := renderTemplate(r.URL, ctx)
	if err != nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: %v \n\n",
			r.ToErrorLine("Error Rendering URL"),
			err,
		)
	}

	resp, err := page.Goto(url)
	if err != nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: %v \n URL: %v \n\n",
			r.ToErrorLine("Error with GOTO"),
			err,
			url,
		)
	}

	if resp == nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: response is nil\n\n",
			r.ToErrorLine("Error with GOTO"),
		)
	}

	for _, assert := range r.Asserts {
		assert.Execute(t, ctx, page, []playwright.Response{resp})
	}

	// Implement the execution logic for the Goto action here
	return page, nil
}

//////
// CLICK
//////

type Click struct {
	Locator string

	Asserts []Assert

	Path string
	Line int
}

func (r *Click) ToFlatString() string {
	return fmt.Sprintf("CLICK %s", r.Locator)
}

func (r *Click) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.Path, r.Line, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.Path, r.Line, msg, r.ToFlatString())
}

func (r *Click) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Click) AddAssert(assert Assert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Click) NeedsResponseOrPage() (bool, bool) {
	// check for new page first
	if r.Asserts != nil || len(r.Asserts) == 0 {
		return false, false
	}

	for _, a := range r.Asserts {
		// if there is a AssertNewPage assert, we need a new page
		_, ok := a.(*AssertNewPage)
		if ok {
			return false, true
		}
	}

	for _, a := range r.Asserts {
		// if there is a AssertResponse assert, we need a response
		_, ok := a.(*AssertStatus)
		if ok {
			return true, false
		}
	}

	return false, false
}

func (r *Click) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	context := page.Context()

	html, err := page.Content()
	if err != nil {
		return nil, err
	}
	prettyOutput := PrettifyHTML(html)

	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("Failed to get locator for %s", r.Locator)
	}

	var responses []playwright.Response
	var requests []playwright.Request
	var newPages []playwright.Page

	context.On("page", func(p playwright.Page) {
		newPages = append(newPages, p)
	})

	context.On("request", func(req playwright.Request) {
		requests = append(requests, req)
	})

	context.On("response", func(resp playwright.Response) {
		responses = append(responses, resp)
	})

	err = locator.Click()
	if err != nil {
		t.Fatalf(
			"\n\n CLICK FAILED \n %s: %v \n\n %s \n\n",
			r.ToErrorLine("Error with CLICK"),
			prettyOutput,
			err,
		)
	}

	// We have to wait for the reponses to be captured
	time.Sleep(50 * time.Millisecond)

	if len(newPages) > 1 {
		t.Fatalf(
			"\n\n CLICK FAILED \n %s: multiple new pages captured\n\n %s \n\n",
			r.ToErrorLine("Multiple new pages with CLICK"),
			prettyOutput,
		)
	}

	newPage := first(newPages)
	if newPage != nil {
		page = newPage

		if strings.Contains(page.URL(), "chrome-error") || page.URL() == "about:blank#blocked" {
			t.Fatalf("%s\n new page failed to load properly\nURL returnes as %s, might be bad URL\n\n", r.ToErrorLine("Error with NEW PAGE"), page.URL())
		}
		err := page.WaitForLoadState(
			playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	for _, assert := range r.Asserts {
		if _, ok := assert.(*AssertNewPage); ok {
			if newPage == nil {
				t.Fatalf("%s\n assertNewPage failed: no new page was opened\n\n", r.ToErrorLine("Error with ASSERT"))
			}
			continue
		}
		err := assert.Execute(t, ctx, page, responses)
		if err != nil {
			t.Fatalf(
				"\n\n ASSERT FAILED \n %s: %v \n\n %s \n\n",
				r.ToErrorLine("Error with ASSERT"),
				err,
				prettyOutput,
			)
		}
	}

	return page, nil
}

func first[T any](slice []T) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	return slice[0]
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

func (r *Input) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.Path, r.Line, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.Path, r.Line, msg, r.ToFlatString())
}

func (r *Input) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Input) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	// Implement the execution logic for the Input action here
	return page, nil
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

func (r *Set) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.Path, r.Line, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.Path, r.Line, msg, r.ToFlatString())
}

func (r *Set) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Set) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	// Implement the execution logic for the Set action here
	return page, nil
}

// /////////////////////
// ASSERTS
// //////////////////////

type Assert interface {
	Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error
}

// //
// ASSERT VISIBLE
// //
type AssertVisible struct {
	Locator string

	Path string
	Line int
}

func (r *AssertVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	html, err := page.Content()
	if err != nil {
		return err
	}

	prettyOutput := PrettifyHTML(html)

	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("%s\n getLocator returned nil for rawLocator: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	visible, err := locator.IsVisible()
	if err != nil {
		t.Fatalf("%s\n failed to check visibility for locator: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
	}
	if !visible {
		t.Fatalf("%s\n assertVisible failed: expected value to be visible for locator: %s\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	return nil
}

func (r *AssertVisible) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected element to be visible for locator %s", r.Path, r.Line, r.Locator)
}

// ///
// ASSERT TEXT EQUALS
// ///

type AssertTextEquals struct {
	Locator string
	Text    string

	Path string
	Line int
}

func (r *AssertTextEquals) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	html, err := page.Content()
	if err != nil {
		return err
	}

	prettyOutput := PrettifyHTML(html)

	expected := normalizeValue(r.Text)
	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("%s\n getLocator returned nil for rawLocator: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	innerText, err := locator.InnerText()
	if err != nil {
		t.Fatalf("%s\n failed to get inner text for locator: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
	}
	if innerText != expected {
		t.Fatalf("%s\n pageContains failed: expected %v at %v found %v: \n\n%s\n\n", r.ToErrorLine(), expected, r.Locator, innerText, prettyOutput)
	}
	return nil
}

func (r *AssertTextEquals) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected text for locator %s", r.Path, r.Line, r.Locator)
}

// ////
// ASSERT STATUS
// ////

type AssertStatus struct {
	Code int

	Path string
	Line int
}

func (r *AssertStatus) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {

	if len(resps) == 0 {
		t.Fatalf("%s\n assertStatus failed: response is nil\n\n", r.ToErrorLine())
	}
	for _, resp := range resps {
		if resp.Status() != r.Code {
			t.Fatalf("%s\n assertStatus failed: expected status code %d, got %d for URL %s\n\n", r.ToErrorLine(), r.Code, resp.Status(), resp.URL())
		}
	}

	return nil
}

func (r *AssertStatus) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected status code %d", r.Path, r.Line, r.Code)
}

// ////
// Just checks the new page is there
// ////
type AssertNewPage struct {
	Path string
	Line int
}

func (r *AssertNewPage) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	return nil
}

func (r *AssertNewPage) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected a new page to be opened", r.Path, r.Line)
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

		split := strings.SplitN(line, " ", 2)
		if len(split) == 0 {
			continue
		}
		second := ""
		if len(split) > 1 {
			second = strings.TrimSpace(split[1])
		}
		first := strings.TrimSpace(split[0])

		switch first {
		case "GOTO":
			flush()

			current = &Goto{
				Path:    path,
				Line:    lineNumber,
				URL:     second,
				Asserts: []Assert{},
			}
		case "CLICK":
			flush()

			current = &Click{
				Path:    path,
				Line:    lineNumber,
				Locator: second,
				Asserts: []Assert{},
			}
		case "INPUT":
			flush()

			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("expected exactly two values for input")
			}

			selector, value := parts[0], parts[1]

			if value == "" {
				return nil, fmt.Errorf("empty input value")
			}

			current = &Input{
				Path:    path,
				Line:    lineNumber,
				Locator: selector,
				Value:   value,
			}
		case "SET":
			flush()

			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("expected exactly two values")
			}

			key, value := parts[0], parts[1]

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
			assertable, ok := current.(interface{ AddAssert(Assert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			statusCode, err := strconv.Atoi(second)
			if err != nil {
				return nil, fmt.Errorf("invalid assert_status: %q", second)
			}
			assertable.AddAssert(&AssertStatus{
				Path: path,
				Line: lineNumber,
				Code: statusCode,
			})

		case "ASSERT_VISIBLE":
			if current == nil {
				return nil, fmt.Errorf("assert_visible before action")
			}
			assertable, ok := current.(interface{ AddAssert(Assert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertVisible{
				Locator: second,
				Path:    path,
				Line:    lineNumber,
			})
		case "ASSERT_TEXT_EQUALS":
			if current == nil {
				return nil, fmt.Errorf("assert_text_equals before action")
			}

			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("expected exactly two values")
			}

			assertable, ok := current.(interface{ AddAssert(Assert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertTextEquals{
				Locator: parts[0],
				Text:    parts[1],
				Path:    path,
				Line:    lineNumber,
			})
		case "ASSERT_NEW_PAGE":
			if current == nil {
				return nil, fmt.Errorf("assert_new_page before action")
			}

			assertable, ok := current.(interface{ AddAssert(Assert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertNewPage{
				Path: path,
				Line: lineNumber,
			})
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

func renderTemplate(input string, ctx map[string]any) (string, error) {
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"server_url": func() string {
				return ctx["server_url"].(string)
			},
			"site": func() string {
				if v, ok := ctx["site"]; ok && v != nil {
					return ctx["site"].(string)
				}
				return ctx["default_site"].(string)
			},
			"jwt": func() string {
				if v, ok := ctx["jwt"]; ok && v != nil {
					return ctx["jwt"].(string)
				}
				return ctx["default_jwt"].(string)
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

func getLocator(expected any, page playwright.Page) playwright.Locator {
	s, ok := expected.(string)
	if !ok {
		return nil
	}

	s = strings.TrimSpace(s)

	switch {
	case strings.HasPrefix(s, "#"):
		// CSS ID: #submit-button
		return page.Locator(s)

	case strings.HasPrefix(s, "."):
		// CSS class: .main-content
		return page.Locator(s)

	case len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`):
		// Text content: "Submit"
		text := strings.Trim(s, `"`)
		return page.GetByText(text)

	default:
		// Normal string value
		return nil
	}

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

func RunTheActionFileAsATest(t *testing.T, root string, path string, context playwright.BrowserContext, serverURL string, siteInformation *db.SiteCreateResult, defaultJwt string) {
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
			"server_url":    serverURL,
			"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
			"default_token": siteInformation.PwdNSToken,
			"default_jwt":   defaultJwt,
		}
		actionFile, err := ParseActionFile(string(actionFileBytes), actionContext, path)
		if err != nil {
			t.Fatalf("failed to parse action file: %v", err)
		}
		for _, action := range actionFile.Actions {
			fmt.Println()
			fmt.Println("=============== " + action.ToFlatString() + "  ================")
			fmt.Println()
			action.PrintAction()
			fmt.Println()
			fmt.Println("- - - - - - - - - - - - - - - - ")
			fmt.Println()
			// default information

			page, err = action.Execute(t, page, actionContext)
			if err != nil {
				t.Fatalf("failed to execute action: %v, action: %s", err, action.ToFlatString())
			}
			if page == nil {
				t.Fatalf("page is nil after executing action: %s", action.ToFlatString())
			}
			fmt.Println("===============================")
		}

	})
}
