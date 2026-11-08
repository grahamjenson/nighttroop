package flow

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseBrowserLines is the actions of a file's lines, each starting one, with the asserts after it
func ParseBrowserLines(lines []Line) ([]Action, error) {
	var actions []Action
	var current Action

	flush := func() {
		if current != nil {
			actions = append(actions, current)
			current = nil
		}
	}
	for _, l := range lines {
		first, second, lineNumber, path := l.First, l.Rest, l.Number, l.Path
		switch first {
		case "GOTO":
			flush()

			current = &Goto{
				TestPath: path,
				TestLine: lineNumber,
				URL:      second,
				Asserts:  []BrowserAssert{},
			}
		case "POST":
			// POST sends a path a JSON body from the page, signed in as it is, e.g. as an app registers itself
			flush()
			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("%s:%d: POST needs a path and a JSON body", path, lineNumber)
			}
			current = &Post{TestPath: path, TestLine: lineNumber, URL: parts[0], Body: parts[1]}
		case "CLICK":
			flush()

			current = &Click{
				TestPath: path,
				TestLine: lineNumber,
				Locator:  second,
				Asserts:  []BrowserAssert{},
			}
		case "CLICK_CORNER", "PRESS":
			flush()

			if second == "" {
				return nil, fmt.Errorf("%s needs a value", first)
			}
			click := &Click{TestPath: path, TestLine: lineNumber, Asserts: []BrowserAssert{}}
			if first == "PRESS" {
				click.Key = second
			} else {
				click.Locator, click.Corner = second, true
			}
			current = click
		case "TYPE":
			flush()

			// TYPE types into a field like INPUT, but takes asserts, for a field that does something once typed in
			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 || parts[1] == "" {
				return nil, fmt.Errorf("TYPE needs a locator and a value")
			}
			current = &Click{TestPath: path, TestLine: lineNumber, Asserts: []BrowserAssert{}, Locator: parts[0], Text: parts[1]}
		case "SCROLL":
			// SCROLL turns the mouse wheel by a number of pixels, down or up when negative, so asserts can check what
			// scrolling shows or hides
			flush()
			pixels, err := strconv.Atoi(second)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: SCROLL needs a number of pixels: %q", path, lineNumber, second)
			}
			current = &Click{TestPath: path, TestLine: lineNumber, Asserts: []BrowserAssert{}, Scroll: &pixels}
		case "BACKGROUND", "FOREGROUND":
			// BACKGROUND hides the page, as a browser does a tab in the background, until FOREGROUND shows it again
			flush()
			current = &Click{TestPath: path, TestLine: lineNumber, Asserts: []BrowserAssert{}, Hidden: new(first == "BACKGROUND")}
		case "CLICK_BACK", "CLICK_FORWARD":
			flush()

			current = &Click{
				TestPath: path,
				TestLine: lineNumber,
				History:  strings.ToLower(strings.TrimPrefix(first, "CLICK_")),
				Asserts:  []BrowserAssert{},
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
				TestPath: path,
				TestLine: lineNumber,
				Locator:  selector,
				Value:    value,
			}
		case "SAVE_COOKIE", "SET_COOKIE":
			flush()

			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("%s:%d: expected a cookie name and a value", path, lineNumber)
			}
			if first == "SAVE_COOKIE" {
				current = &SaveCookie{TestPath: path, TestLine: lineNumber, Name: parts[0], Key: parts[1]}
			} else {
				current = &SetCookie{TestPath: path, TestLine: lineNumber, Name: parts[0], Value: parts[1]}
			}
		case "WINDOW":
			// WINDOW sets the window's size in pixels, e.g. WINDOW 390x844 for a phone, for the pages after it
			flush()
			var width, height int
			if _, err := fmt.Sscanf(second, "%dx%d", &width, &height); err != nil {
				return nil, fmt.Errorf("%s:%d: WINDOW needs a size like 390x844: %q", path, lineNumber, second)
			}
			current = &Window{TestPath: path, TestLine: lineNumber, Width: width, Height: height}
		case "USER_AGENT":
			// USER_AGENT sends the pages' requests with a User-Agent, as another app would, until one with none
			flush()
			current = &UserAgent{TestPath: path, TestLine: lineNumber, Value: second}
		case "SAVE_TEXT", "SAVE_VALUE":
			// SAVE_TEXT keeps an element's text as a value for the steps after it, and SAVE_VALUE an input's value,
			// though it is hidden, as a TOTP secret only shown as a QR code is
			flush()
			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("%s:%d: expected a locator and a key", path, lineNumber)
			}
			current = &SaveText{TestPath: path, TestLine: lineNumber, Locator: parts[0], Key: parts[1], Value: first == "SAVE_VALUE"}
		case "CLEAR_COOKIE":
			flush()

			current = &ClearCookie{TestPath: path, TestLine: lineNumber, Name: second}
		case "RUN_BACKGROUND_JOBS":
			flush()
			// Asserts after it check the page once the jobs have run, e.g. that it updated itself live
			current = &RunBackgroundJobs{TestPath: path, TestLine: lineNumber}
		// Adding screenshot action
		case "SCREENSHOT":
			if current == nil {
				return nil, fmt.Errorf("screenshot before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&Screenshot{
				TestPath: path,
				TestLine: lineNumber,
				Name:     second,
			})
		case "ASSERT_STATUS":
			if current == nil {
				return nil, fmt.Errorf("assert_status before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			statusCode, err := strconv.Atoi(second)
			if err != nil {
				return nil, fmt.Errorf("invalid assert_status: %q", second)
			}
			assertable.AddAssert(&BrowserAssertStatus{
				TestPath: path,
				TestLine: lineNumber,
				Code:     statusCode,
			})

		case "ASSERT_VISIBLE", "ASSERT_DRAWN":
			if current == nil {
				return nil, fmt.Errorf("assert_visible before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertVisible{
				Locator:  second,
				Drawn:    first == "ASSERT_DRAWN",
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_NOT_VISIBLE":
			if current == nil {
				return nil, fmt.Errorf("assert_not_visible before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertNotVisible{
				Locator:  second,
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_TEXT_VISIBLE":
			if current == nil {
				return nil, fmt.Errorf("assert_text_visible before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertTextVisible{
				Text:     second,
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_TEXT_EQUALS":
			if current == nil {
				return nil, fmt.Errorf("assert_text_equals before action")
			}

			parts := strings.SplitN(second, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("expected exactly two values")
			}

			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertTextEquals{
				Locator:  parts[0],
				Text:     parts[1],
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_NEW_PAGE":
			if current == nil {
				return nil, fmt.Errorf("assert_new_page before action")
			}

			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertNewPage{
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_NEW_URL":
			if current == nil {
				return nil, fmt.Errorf("assert_new_url before action")
			}

			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertNewURL{
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_PATH":
			if current == nil {
				return nil, fmt.Errorf("assert_path before action")
			}

			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertPath{
				URLPath:  second,
				TestPath: path,
				TestLine: lineNumber,
			})
		case "ASSERT_FITS":
			// ASSERT_FITS is the whole page fitting in the window, so nothing on it needs scrolling to see
			if current == nil {
				return nil, fmt.Errorf("assert_fits before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertFits{TestPath: path, TestLine: lineNumber})
		case "ASSERT_COUNT":
			// ASSERT_COUNT is how many elements the locator finds are visible, counted at once, as for toasts that
			// time out, which a wait would let go
			if current == nil {
				return nil, fmt.Errorf("assert_count before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			parts := strings.SplitN(second, " ", 2)
			count, err := strconv.Atoi(parts[len(parts)-1])
			if len(parts) != 2 || err != nil {
				return nil, fmt.Errorf("%s:%d: ASSERT_COUNT needs a locator and a number: %q", path, lineNumber, second)
			}
			assertable.AddAssert(&AssertCount{Locator: parts[0], Count: count, TestPath: path, TestLine: lineNumber})
		case "ASSERT_TITLE":
			if current == nil {
				return nil, fmt.Errorf("assert_title before action")
			}

			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertTitle{
				Title:    second,
				TestPath: path,
				TestLine: lineNumber,
			})
		default:
			return nil, fmt.Errorf("unknown line: %q", l.Text)
		}

	}

	flush()

	return actions, nil
}
