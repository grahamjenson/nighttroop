package display

import (
	"context"
	"strings"
)

// separators are each number format's thousands separator and decimal mark, its space one that does not break
var separators = map[string][2]string{"1,234.5": {",", "."}, "1.234,5": {".", ","}, "1 234,5": {" ", ","}}

type numberFormatKey struct{}

// WithNumberFormat is the context with its numbers written in the number format, as 1,234.5, 1.234,5 or 1 234,5
func WithNumberFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, numberFormatKey{}, format)
}

// Number is a number written with a "." decimal point and no thousands separator, as strconv writes it, in the
// context's number format, or 1,234.5 for none: "-1234.5" is "-1,234.5", "-1.234,5" or "-1 234,5"
func Number(ctx context.Context, number string) string {
	format, _ := ctx.Value(numberFormatKey{}).(string)
	marks, ok := separators[format]
	if !ok {
		marks = separators["1,234.5"]
	}
	sign, digits := "", number
	if strings.HasPrefix(digits, "-") {
		sign, digits = "-", digits[1:]
	}
	whole, fraction, hasFraction := strings.Cut(digits, ".")
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + marks[0] + whole[i:]
	}
	if hasFraction {
		return sign + whole + marks[1] + fraction
	}
	return sign + whole
}
