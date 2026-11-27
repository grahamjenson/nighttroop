package flow

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

// parseBrowser is the actions of the file at path in fsys
func parseBrowser(fsys fs.FS, path string) ([]Action, error) {
	lines, err := ReadLines(fsys, path)
	if err != nil {
		return nil, err
	}
	return ParseBrowserLines(lines)
}

// Each browser step is parsed to its action, with the asserts after it. A file can INCLUDE steps/login.steps.
func TestParseBrowserSteps(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		actions  int
		check    func(t *testing.T, file []Action)
	}{
		{"INCLUDE runs another file's steps", "INCLUDE steps/login.steps\nASSERT_VISIBLE #home\n\nCLICK #next\n", 3, func(t *testing.T, file []Action) {
			require.IsType(t, &Goto{}, file[0])
			// An assert after the INCLUDE checks the page the included file's last step left
			login := file[1].(*Click)
			require.Equal(t, "#login", login.Locator)
			require.Len(t, login.Asserts, 2)
			require.Equal(t, "#next", file[2].(*Click).Locator)
		}},
		{"SCROLL turns the wheel by pixels", "GOTO /\nSCROLL 400\nASSERT_NOT_VISIBLE #header\nSCROLL -100\n", 3, func(t *testing.T, file []Action) {
			down := file[1].(*Click)
			require.Equal(t, 400, *down.Scroll)
			require.Len(t, down.Asserts, 1)
			require.Equal(t, "SCROLL -100", file[2].ToFlatString())
		}},
		{"HOLD holds a path's requests until released", "GOTO /\nHOLD /refresh\nRUN_BACKGROUND_JOBS\nASSERT_HELD /refresh\nRELEASE /refresh\nASSERT_VISIBLE #done\n", 4, func(t *testing.T, file []Action) {
			require.Equal(t, &Hold{Path: "/refresh", TestPath: "site/test.bdd", TestLine: 2}, file[1])
			require.Equal(t, []BrowserAssert{&AssertHeld{Path: "/refresh", TestPath: "site/test.bdd", TestLine: 4}}, file[2].(*RunBackgroundJobs).Asserts)
			release := file[3].(*Hold)
			require.Equal(t, "RELEASE /refresh", release.ToFlatString())
			require.Len(t, release.Asserts, 1)
		}},
		{"WAIT_VISIBLE waits seconds for a locator", "GOTO /\nWAIT_VISIBLE 30 #badge:text-matches(\"^No reading$\")\nASSERT_VISIBLE #grey\n", 2, func(t *testing.T, file []Action) {
			wait := file[1].(*WaitVisible)
			require.Equal(t, 30, wait.Seconds)
			require.Equal(t, `#badge:text-matches("^No reading$")`, wait.Locator)
			require.Len(t, wait.Asserts, 1)
		}},
		{"HOVER points at a share across", "GOTO /\nHOVER #graph svg\nASSERT_TEXT_VISIBLE Basal\nHOVER 95% #graph svg\n", 3, func(t *testing.T, file []Action) {
			// Without a share, it points at the middle, and asserts after it check what pointing there shows
			middle := file[1].(*Click)
			require.Equal(t, "#graph svg", middle.Locator)
			require.Equal(t, 0.5, *middle.Hover)
			require.Len(t, middle.Asserts, 1)
			near := file[2].(*Click)
			require.Equal(t, "#graph svg", near.Locator)
			require.InDelta(t, 0.95, *near.Hover, 1e-9)
			require.Equal(t, "HOVER 95% #graph svg", near.ToFlatString())
		}},
		{"ASSERT_TITLE keeps the whole title", "GOTO /data/glucose\nASSERT_TITLE Glucose Readings · Nighttroop\n", 1, func(t *testing.T, file []Action) {
			asserts := file[0].(*Goto).Asserts
			require.Len(t, asserts, 1)
			require.Equal(t, "Glucose Readings · Nighttroop", asserts[0].(*AssertTitle).Title)
		}},
		{"SAVE_VALUE keeps an input's value", "GOTO /\nSAVE_TEXT #password password\nSAVE_VALUE #totp-secret totp_secret\n", 3, func(t *testing.T, file []Action) {
			require.False(t, file[1].(*SaveText).Value)
			require.Equal(t, &SaveText{TestPath: "site/test.bdd", TestLine: 3, Locator: "#totp-secret", Key: "totp_secret", Value: true}, file[2])
			require.Equal(t, "SAVE_VALUE #totp-secret totp_secret", file[2].ToFlatString())
		}},
		{"ASSERT_COUNT takes a locator and a number", "GOTO /\nASSERT_COUNT #toaster>[data-tui-toast] 3\nASSERT_COUNT #table tbody tr 30\n", 1, func(t *testing.T, file []Action) {
			count := file[0].(*Goto).Asserts[0].(*AssertCount)
			require.Equal(t, "#toaster>[data-tui-toast]", count.Locator)
			require.Equal(t, 3, count.Count)
			// The number is the last word, so the locator can have spaces
			rows := file[0].(*Goto).Asserts[1].(*AssertCount)
			require.Equal(t, "#table tbody tr", rows.Locator)
			require.Equal(t, 30, rows.Count)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"steps/login.steps": {Data: []byte("GOTO /login\nCLICK #login\nASSERT_STATUS 200\n")},
				"site/test.bdd":     {Data: []byte(test.contents)},
			}
			file, err := parseBrowser(fsys, "site/test.bdd")
			require.NoError(t, err)
			require.Len(t, file, test.actions)
			test.check(t, file)
		})
	}
}

// A browser file's lines that cannot be parsed are refused, saying the file and line where it can
func TestParseBrowserErrors(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		err      string
	}{
		{"SCROLL without pixels", "GOTO /\nSCROLL down\n", "site/bad.bdd:2: SCROLL needs a number of pixels"},
		{"HOVER past the right", "GOTO /\nHOVER 150% #graph\n", "site/bad.bdd:2: HOVER's share across is a percentage from 0 to 100"},
		{"HOLD without a path", "GOTO /\nHOLD\n", "site/bad.bdd:2: HOLD needs a path"},
		{"WAIT_VISIBLE without seconds", "GOTO /\nWAIT_VISIBLE #badge\n", "site/bad.bdd:2: WAIT_VISIBLE needs 1 to 60 seconds and a locator"},
		{"WAIT_VISIBLE for no seconds", "GOTO /\nWAIT_VISIBLE 0 #badge\n", "site/bad.bdd:2: WAIT_VISIBLE needs 1 to 60 seconds and a locator"},
		{"WAIT_VISIBLE over a minute", "GOTO /\nWAIT_VISIBLE 61 #badge\n", "site/bad.bdd:2: WAIT_VISIBLE needs 1 to 60 seconds and a locator"},
		{"WAIT_VISIBLE without a locator", "GOTO /\nWAIT_VISIBLE 30\n", "site/bad.bdd:2: WAIT_VISIBLE needs 1 to 60 seconds and a locator"},
		{"ASSERT_TITLE before an action", "ASSERT_TITLE Nighttroop\n", "assert_title before action"},
		{"SAVE_VALUE without a key", "GOTO /\nSAVE_VALUE #totp-secret\n", "site/bad.bdd:2: expected a locator and a key"},
		// An assert that the action before it would not check is refused, as it would pass whatever the page did
		{"ASSERT_NEW_URL after GOTO", "GOTO /\nASSERT_NEW_URL\n", "site/bad.bdd:2: ASSERT_NEW_URL checks a click, not GOTO /"},
		{"ASSERT_NEW_PAGE after RUN_BACKGROUND_JOBS", "GOTO /\nRUN_BACKGROUND_JOBS\nASSERT_NEW_PAGE\n", "site/bad.bdd:3: ASSERT_NEW_PAGE checks a click, not RUN_BACKGROUND_JOBS"},
		{"an assert after MIDDAY_TIMEZONE", "MIDDAY_TIMEZONE\nASSERT_VISIBLE #summary\n", "site/bad.bdd:2: ASSERT_VISIBLE cannot follow MIDDAY_TIMEZONE, which takes no asserts"},
		{"ASSERT_COUNT without a number", "GOTO /\nASSERT_COUNT #toaster three\n", "site/bad.bdd:2: ASSERT_COUNT needs a locator and a number"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseBrowser(fstest.MapFS{"site/bad.bdd": {Data: []byte(test.contents)}}, "site/bad.bdd")
			require.ErrorContains(t, err, test.err)
		})
	}
}
