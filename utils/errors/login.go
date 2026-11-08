package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// LocalStack is where in Nighttroop's code its caller was called from, e.g. the handler a login failed in, for its log
func LocalStack() string {
	const depth = 32
	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs)

	frames := runtime.CallersFrames(pcs[:n])

	var out []string

	for {
		f, more := frames.Next()

		if strings.Contains(f.File, "nighttroop") {
			out = append(out, fmt.Sprintf("%s:%d", f.File, f.Line))
		}

		if !more {
			break
		}
	}

	return strings.Join(out, " → ")
}
