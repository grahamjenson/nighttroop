package flow

import (
	"fmt"
	"strconv"
)

// ParseAssert is the assert on an ASSERT_* line, and false when the line is not one
func ParseAssert(first, second, path string, lineNumber int) (HTTPAssert, bool, error) {
	switch first {
	case "ASSERT_STATUS":
		statusCode, err := strconv.Atoi(second)
		if err != nil {
			return nil, true, fmt.Errorf("invalid assert_status: %q", second)
		}
		return HTTPAssertStatus{Path: path, Line: lineNumber, Code: statusCode}, true, nil

	case "ASSERT_EQUALS", "ASSERT_CONTAINS":
		expected, actual, err := parseAssertArgs(second)
		if err != nil {
			return nil, true, fmt.Errorf("%s:%d: invalid %s: %w", path, lineNumber, first, err)
		}
		if first == "ASSERT_EQUALS" {
			return AssertEquals{Path: path, Line: lineNumber, Expected: expected, Actual: actual}, true, nil
		}
		return AssertContains{Path: path, Line: lineNumber, Expected: expected, Actual: actual}, true, nil

	case "ASSERT_EXISTS":
		actual, err := strconv.Unquote(second)
		if err != nil {
			return nil, true, fmt.Errorf("%s:%d: invalid %s: value %q must be a quoted string", path, lineNumber, first, second)
		}
		return AssertExists{Path: path, Line: lineNumber, Actual: actual}, true, nil

	case "ASSERT_LENGTH":
		expected, actual, err := parseAssertArgs(second)
		length, isNumber := expected.(float64)
		if err == nil && (!isNumber || length != float64(int(length))) {
			err = fmt.Errorf("expected length %v must be a whole number", expected)
		}
		if err != nil {
			return nil, true, fmt.Errorf("%s:%d: invalid %s: %w", path, lineNumber, first, err)
		}
		return AssertLength{Path: path, Line: lineNumber, Expected: int(length), Actual: actual}, true, nil
	}
	return nil, false, nil
}
