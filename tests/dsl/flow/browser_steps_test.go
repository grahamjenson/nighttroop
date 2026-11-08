package flow

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestIncludeRunsAnotherFilesSteps(t *testing.T) {
	fsys := fstest.MapFS{
		"steps/login.steps": {Data: []byte("GOTO /login\nCLICK #login\nASSERT_STATUS 200\n")},
		"site/home.bdd":     {Data: []byte("INCLUDE steps/login.steps\nASSERT_VISIBLE #home\n\nCLICK #next\n")},
		"site/missing.bdd":  {Data: []byte("CLICK #start\nINCLUDE steps/missing.steps\n")},
	}

	file, err := parseBrowser(fsys, "site/home.bdd")
	require.NoError(t, err)
	require.Len(t, file, 3)
	require.IsType(t, &Goto{}, file[0])
	// An assert after the INCLUDE checks the page the included file's last step left
	login := file[1].(*Click)
	require.Equal(t, "#login", login.Locator)
	require.Len(t, login.Asserts, 2)
	require.Equal(t, "#next", file[2].(*Click).Locator)

	_, err = parseBrowser(fsys, "site/missing.bdd")
	require.ErrorContains(t, err, "site/missing.bdd:2: INCLUDE steps/missing.steps")
}

func TestScrollTurnsTheWheelByPixels(t *testing.T) {
	fsys := fstest.MapFS{
		"site/scroll.bdd": {Data: []byte("GOTO /\nSCROLL 400\nASSERT_NOT_VISIBLE #header\nSCROLL -100\n")},
		"site/bad.bdd":    {Data: []byte("GOTO /\nSCROLL down\n")},
	}

	file, err := parseBrowser(fsys, "site/scroll.bdd")
	require.NoError(t, err)
	require.Len(t, file, 3)
	down := file[1].(*Click)
	require.Equal(t, 400, *down.Scroll)
	require.Len(t, down.Asserts, 1)
	require.Equal(t, "SCROLL -100", file[2].ToFlatString())

	_, err = parseBrowser(fsys, "site/bad.bdd")
	require.ErrorContains(t, err, "site/bad.bdd:2: SCROLL needs a number of pixels")
}

func TestHoverPointsAtAShareAcross(t *testing.T) {
	fsys := fstest.MapFS{
		"site/hover.bdd": {Data: []byte("GOTO /\nHOVER #graph svg\nASSERT_TEXT_VISIBLE Basal\nHOVER 95% #graph svg\n")},
		"site/bad.bdd":   {Data: []byte("GOTO /\nHOVER 150% #graph\n")},
	}

	file, err := parseBrowser(fsys, "site/hover.bdd")
	require.NoError(t, err)
	require.Len(t, file, 3)
	// Without a share, it points at the middle, and asserts after it check what pointing there shows
	middle := file[1].(*Click)
	require.Equal(t, "#graph svg", middle.Locator)
	require.Equal(t, 0.5, *middle.Hover)
	require.Len(t, middle.Asserts, 1)
	near := file[2].(*Click)
	require.Equal(t, "#graph svg", near.Locator)
	require.InDelta(t, 0.95, *near.Hover, 1e-9)
	require.Equal(t, "HOVER 95% #graph svg", near.ToFlatString())

	_, err = parseBrowser(fsys, "site/bad.bdd")
	require.ErrorContains(t, err, "site/bad.bdd:2: HOVER's share across is a percentage from 0 to 100")
}

func TestTitleAssertKeepsTheWholeTitle(t *testing.T) {
	fsys := fstest.MapFS{
		"site/title.bdd": {Data: []byte("GOTO /data/glucose\nASSERT_TITLE Glucose Readings · Nighttroop\n")},
		"site/bad.bdd":   {Data: []byte("ASSERT_TITLE Nighttroop\n")},
	}

	file, err := parseBrowser(fsys, "site/title.bdd")
	require.NoError(t, err)
	require.Len(t, file, 1)
	asserts := file[0].(*Goto).Asserts
	require.Len(t, asserts, 1)
	require.Equal(t, "Glucose Readings · Nighttroop", asserts[0].(*AssertTitle).Title)

	_, err = parseBrowser(fsys, "site/bad.bdd")
	require.ErrorContains(t, err, "assert_title before action")
}

// parseBrowser is the actions of the file at path in fsys
func parseBrowser(fsys fs.FS, path string) ([]Action, error) {
	lines, err := ReadLines(fsys, path)
	if err != nil {
		return nil, err
	}
	return ParseBrowserLines(lines)
}

func TestSaveValueKeepsAnInputsValue(t *testing.T) {
	fsys := fstest.MapFS{
		"site/save.bdd": {Data: []byte("GOTO /\nSAVE_TEXT #password password\nSAVE_VALUE #totp-secret totp_secret\n")},
		"site/bad.bdd":  {Data: []byte("GOTO /\nSAVE_VALUE #totp-secret\n")},
	}

	file, err := parseBrowser(fsys, "site/save.bdd")
	require.NoError(t, err)
	require.Len(t, file, 3)
	require.False(t, file[1].(*SaveText).Value)
	require.Equal(t, &SaveText{TestPath: "site/save.bdd", TestLine: 3, Locator: "#totp-secret", Key: "totp_secret", Value: true}, file[2])
	require.Equal(t, "SAVE_VALUE #totp-secret totp_secret", file[2].ToFlatString())

	_, err = parseBrowser(fsys, "site/bad.bdd")
	require.ErrorContains(t, err, "site/bad.bdd:2: expected a locator and a key")
}

func TestCountAssertTakesALocatorAndANumber(t *testing.T) {
	fsys := fstest.MapFS{
		"site/count.bdd": {Data: []byte("GOTO /\nASSERT_COUNT #toaster>[data-tui-toast] 3\n")},
		"site/bad.bdd":   {Data: []byte("GOTO /\nASSERT_COUNT #toaster three\n")},
	}

	file, err := parseBrowser(fsys, "site/count.bdd")
	require.NoError(t, err)
	count := file[0].(*Goto).Asserts[0].(*AssertCount)
	require.Equal(t, "#toaster>[data-tui-toast]", count.Locator)
	require.Equal(t, 3, count.Count)

	_, err = parseBrowser(fsys, "site/bad.bdd")
	require.ErrorContains(t, err, "site/bad.bdd:2: ASSERT_COUNT needs a locator and a number")
}
