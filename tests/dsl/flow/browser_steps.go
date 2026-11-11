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
		case "CLICK_ASKS":
			// CLICK_ASKS <locator> clicks what asks to be confirmed and leaves the site's confirm dialog open, to check what
			// it asks, then CLICK #confirm--ok or #confirm--cancel answers it; a plain CLICK confirms it
			flush()
			current = &Click{TestPath: path, TestLine: lineNumber, Locator: second, Asks: true, Asserts: []BrowserAssert{}}
		case "CLICK_QUICKLY":
			// CLICK_QUICKLY <selector>, <selector>, ... clicks each CSS selector's element straight after the last, not
			// waiting for what a click asks for, as a person clicking around quickly does, so each request replaces the
			// one before
			flush()
			if second == "" {
				return nil, fmt.Errorf("%s:%d: CLICK_QUICKLY needs locators", path, lineNumber)
			}
			current = &Click{TestPath: path, TestLine: lineNumber, Quickly: strings.Split(second, ", "), Asserts: []BrowserAssert{}}
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
		case "HOVER":
			// HOVER [x%] <locator> moves the mouse over an element, at its centre or x% of the way across it, so asserts
			// can check what pointing at it shows, such as a graph's tooltip at a time
			flush()
			across, locator := 0.5, second
			if share, rest, ok := strings.Cut(second, "% "); ok {
				percent, err := strconv.ParseFloat(share, 64)
				if err != nil || percent < 0 || percent > 100 {
					return nil, fmt.Errorf("%s:%d: HOVER's share across is a percentage from 0 to 100", path, lineNumber)
				}
				across, locator = percent/100, rest
			}
			if locator == "" {
				return nil, fmt.Errorf("%s:%d: HOVER needs a locator", path, lineNumber)
			}
			current = &Click{TestPath: path, TestLine: lineNumber, Asserts: []BrowserAssert{}, Locator: locator, Hover: &across}
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
		case "MEDIA":
			// MEDIA print shows the pages after it as they print, and MEDIA screen as a screen shows them again
			flush()
			if second != "print" && second != "screen" {
				return nil, fmt.Errorf("%s:%d: MEDIA needs print or screen: %q", path, lineNumber, second)
			}
			current = &Media{TestPath: path, TestLine: lineNumber, Media: second}
		case "VIRTUAL_AUTHENTICATOR":
			// VIRTUAL_AUTHENTICATOR gives the page a passkey authenticator, as a phone's, that verifies the person itself
			flush()
			current = &VirtualAuthenticator{TestPath: path, TestLine: lineNumber}
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
		case "HOLD", "RELEASE":
			// HOLD <path> holds the page's requests to path, unsent, until RELEASE <path> sends them and those after, as
			// a slow server would, so steps between can act while one is on its way
			flush()
			if second == "" {
				return nil, fmt.Errorf("%s:%d: %s needs a path", path, lineNumber, first)
			}
			current = &Hold{TestPath: path, TestLine: lineNumber, Path: second, Release: first == "RELEASE"}
		case "RUN_BACKGROUND_JOBS":
			flush()
			// Asserts after it check the page once the jobs have run, e.g. that it updated itself live
			current = &RunBackgroundJobs{TestPath: path, TestLine: lineNumber}
		case "WAIT_VISIBLE":
			// WAIT_VISIBLE <seconds> <locator> waits up to that many seconds, longer than an assert's 5, for what a timer
			// on the page shows, e.g. a reload when a reading turns stale; asserts after it check the page then
			flush()
			parts := strings.SplitN(second, " ", 2)
			seconds, err := strconv.Atoi(parts[0])
			if len(parts) != 2 || err != nil || seconds < 1 || seconds > 60 {
				return nil, fmt.Errorf("%s:%d: WAIT_VISIBLE needs 1 to 60 seconds and a locator", path, lineNumber)
			}
			current = &WaitVisible{TestPath: path, TestLine: lineNumber, Seconds: seconds, Locator: parts[1]}
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

		case "ASSERT_VISIBLE", "ASSERT_DRAWN", "ASSERT_CONTAINED":
			if current == nil {
				return nil, fmt.Errorf("assert_visible before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertVisible{
				Locator:   second,
				Drawn:     first == "ASSERT_DRAWN",
				Contained: first == "ASSERT_CONTAINED",
				TestPath:  path,
				TestLine:  lineNumber,
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
		case "ASSERT_HELD":
			// ASSERT_HELD <path> waits for the page to send a request to the path HOLD holds
			if current == nil {
				return nil, fmt.Errorf("assert_held before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertHeld{Path: second, TestPath: path, TestLine: lineNumber})
		case "ASSERT_NO_PAGE_ERRORS":
			// ASSERT_NO_PAGE_ERRORS is the page having thrown no uncaught error, nor htmx logged one, since the flow went to it
			if current == nil {
				return nil, fmt.Errorf("assert_no_page_errors before action")
			}
			assertable, ok := current.(interface{ AddAssert(BrowserAssert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			assertable.AddAssert(&AssertNoPageErrors{TestPath: path, TestLine: lineNumber})
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
			// The number is the last word, so the locator can have spaces, as one inside a table does
			split := strings.LastIndex(second, " ")
			count, err := strconv.Atoi(second[split+1:])
			if split <= 0 || err != nil {
				return nil, fmt.Errorf("%s:%d: ASSERT_COUNT needs a locator and a number: %q", path, lineNumber, second)
			}
			assertable.AddAssert(&AssertCount{Locator: second[:split], Count: count, TestPath: path, TestLine: lineNumber})
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
