package flow

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

// A flow's steps, and the server's logs while it ran, are kept rather than printed, and shown only when it fails, so a
// failing run shows the flows that failed rather than every step and log of every flow.

// color is whether the report is coloured, as GitHub Actions and terminals show it, unless NO_COLOR is set
var color = os.Getenv("NO_COLOR") == ""

// style wraps text in the ANSI codes, e.g. "1;31" for bold red, when coloured
func style(codes, text string) string {
	if !color {
		return text
	}
	return "\x1b[" + codes + "m" + text + "\x1b[0m"
}

func bold(text string) string  { return style("1", text) }
func red(text string) string   { return style("1;31", text) }
func green(text string) string { return style("1;32", text) }
func dim(text string) string   { return style("2", text) }

// trails are the steps each running flow took, by its test
var trails sync.Map

// logf adds a line to the steps the test's flow took, shown if it fails
func logf(t testing.TB, format string, args ...any) {
	if trail, ok := trails.Load(t); ok {
		trail.(*strings.Builder).WriteString(fmt.Sprintf(format, args...) + "\n")
	}
}

// ServerLogs is where the test server's logs go, to be shown with a flow that fails, rather than printed
var ServerLogs = &serverLogs{}

// maxServerLogLines is how many of the server's last log lines are kept, and the most shown with a failure
const (
	maxServerLogLines = 20000
	shownServerLogs   = 200
)

type serverLogs struct {
	mu      sync.Mutex
	lines   []string
	dropped int // lines no longer kept, so a mark is still a position in all the lines written
}

func (l *serverLogs) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, strings.Split(strings.TrimRight(string(p), "\n"), "\n")...)
	if over := len(l.lines) - maxServerLogLines; over > 0 {
		l.lines, l.dropped = l.lines[over:], l.dropped+over
	}
	return len(p), nil
}

// mark is where the logs are up to
func (l *serverLogs) mark() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.dropped + len(l.lines)
}

// since are the last shownServerLogs lines logged since the mark, warnings and errors highlighted
func (l *serverLogs) since(mark int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	lines := l.lines[max(mark-l.dropped, len(l.lines)-shownServerLogs, 0):]
	var out strings.Builder
	for _, line := range lines {
		switch {
		case strings.Contains(line, "level=ERROR"):
			line = red(line)
		case strings.Contains(line, "level=WARN"):
			line = style("33", line)
		default:
			line = dim(line)
		}
		out.WriteString(line + "\n")
	}
	return out.String()
}

// report keeps the flow's steps and the server's logs while it runs, and shows them only if the test fails
func report(t *testing.T, path string) {
	trail, from := &strings.Builder{}, ServerLogs.mark()
	trails.Store(t, trail)
	t.Cleanup(func() {
		trails.Delete(t)
		if !t.Failed() {
			return
		}
		t.Logf("\n%s\n\n%s\n%s\n%s\n%s",
			red("✗ FLOW FAILED: "+path),
			bold("Steps, the last the one that failed:"), trail.String(),
			bold("Server logs while it ran (other flows running at the same time log here too):"), ServerLogs.since(from))
	})
}

// failure is an assert's failure, the expected and actual values first, then what it read them from
func failure(title, file string, line int, expected, actual any, message string) string {
	return fmt.Sprintf("\n%s %s\n  %s %s\n  %s %s\n\n%s", red("✗ "+title), bold(fmt.Sprintf("%s:%d", file, line)),
		bold("expected:"), green(fmt.Sprintf("%#v", expected)), bold("actual:  "), red(fmt.Sprintf("%#v", actual)), dim(message))
}
