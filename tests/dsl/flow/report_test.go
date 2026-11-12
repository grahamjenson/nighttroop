package flow

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// A failing flow shows the server's logs since it started, at most the last shownServerLogs, its warnings and errors
// highlighted, and none from before it
func TestServerLogsSinceAMark(t *testing.T) {
	logs := &serverLogs{}
	fmt.Fprintln(logs, "level=INFO msg=before")
	mark := logs.mark()
	fmt.Fprint(logs, "level=INFO msg=during\nlevel=WARN msg=slow\nlevel=ERROR msg=broke\n")
	got := logs.since(mark)
	require.NotContains(t, got, "before")
	require.Contains(t, got, dim("level=INFO msg=during"))
	require.Contains(t, got, style("33", "level=WARN msg=slow"))
	require.Contains(t, got, red("level=ERROR msg=broke"))

	for i := range shownServerLogs + 10 {
		fmt.Fprintf(logs, "level=INFO msg=line%d\n", i)
	}
	got = logs.since(mark)
	require.Equal(t, shownServerLogs, strings.Count(got, "\n"), "only the last lines are shown")
	require.Contains(t, got, fmt.Sprintf("msg=line%d", shownServerLogs+9))

	for range maxServerLogLines {
		fmt.Fprintln(logs, "level=INFO msg=more")
	}
	require.Equal(t, maxServerLogLines, len(logs.lines), "the oldest lines are dropped")
	require.Equal(t, logs.dropped+maxServerLogLines, logs.mark(), "a mark still counts every line written")
}

// An assert's failure leads with where it is and the expected and actual values, before the output it read them from
func TestFailureLeadsWithExpectedAndActual(t *testing.T) {
	color = false
	t.Cleanup(func() { color = true })
	got := failure("ASSERT FAILED", "aaps/sync_v3.http", 52, 0.4, nil, "output: {...}")
	require.Equal(t, "\n✗ ASSERT FAILED aaps/sync_v3.http:52\n  expected: 0.4\n  actual:   <nil>\n\noutput: {...}", got)
}
