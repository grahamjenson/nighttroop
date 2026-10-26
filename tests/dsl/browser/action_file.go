package browser

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

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
				TestPath: path,
				TestLine: lineNumber,
				URL:      second,
				Asserts:  []Assert{},
			}
		case "CLICK":
			flush()

			current = &Click{
				TestPath: path,
				TestLine: lineNumber,
				Locator:  second,
				Asserts:  []Assert{},
			}
		case "CLICK_BACK", "CLICK_FORWARD":
			flush()

			current = &Click{
				TestPath: path,
				TestLine: lineNumber,
				History:  strings.ToLower(strings.TrimPrefix(first, "CLICK_")),
				Asserts:  []Assert{},
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
				TestPath: path,
				TestLine: lineNumber,
				Key:      key,
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
		case "CLEAR_COOKIE":
			flush()

			current = &ClearCookie{TestPath: path, TestLine: lineNumber, Name: second}
		case "RUN_BACKGROUND_JOBS":
			flush()
			actions = append(actions, &RunBackgroundJobs{TestPath: path, TestLine: lineNumber})
		// Adding screenshot action
		case "SCREENSHOT":
			if current == nil {
				return nil, fmt.Errorf("screenshot before action")
			}
			assertable, ok := current.(interface{ AddAssert(Assert) })
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
			assertable, ok := current.(interface{ AddAssert(Assert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}
			statusCode, err := strconv.Atoi(second)
			if err != nil {
				return nil, fmt.Errorf("invalid assert_status: %q", second)
			}
			assertable.AddAssert(&AssertStatus{
				TestPath: path,
				TestLine: lineNumber,
				Code:     statusCode,
			})

		case "ASSERT_VISIBLE", "ASSERT_DRAWN":
			if current == nil {
				return nil, fmt.Errorf("assert_visible before action")
			}
			assertable, ok := current.(interface{ AddAssert(Assert) })
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
			assertable, ok := current.(interface{ AddAssert(Assert) })
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
			assertable, ok := current.(interface{ AddAssert(Assert) })
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

			assertable, ok := current.(interface{ AddAssert(Assert) })
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

			assertable, ok := current.(interface{ AddAssert(Assert) })
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

			assertable, ok := current.(interface{ AddAssert(Assert) })
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

			assertable, ok := current.(interface{ AddAssert(Assert) })
			if !ok {
				return nil, fmt.Errorf("current action does not support asserts")
			}

			assertable.AddAssert(&AssertPath{
				URLPath:  second,
				TestPath: path,
				TestLine: lineNumber,
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
