package browser

import (
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

	file, err := ParseActionFile(fsys, "site/home.bdd")
	require.NoError(t, err)
	require.Equal(t, "home", file.Name)
	require.Len(t, file.Actions, 3)
	require.IsType(t, &Goto{}, file.Actions[0])
	// An assert after the INCLUDE checks the page the included file's last step left
	login := file.Actions[1].(*Click)
	require.Equal(t, "#login", login.Locator)
	require.Len(t, login.Asserts, 2)
	require.Equal(t, "#next", file.Actions[2].(*Click).Locator)

	_, err = ParseActionFile(fsys, "site/missing.bdd")
	require.ErrorContains(t, err, "site/missing.bdd:2: INCLUDE steps/missing.steps")
}

func TestScrollTurnsTheWheelByPixels(t *testing.T) {
	fsys := fstest.MapFS{
		"site/scroll.bdd": {Data: []byte("GOTO /\nSCROLL 400\nASSERT_NOT_VISIBLE #header\nSCROLL -100\n")},
		"site/bad.bdd":    {Data: []byte("GOTO /\nSCROLL down\n")},
	}

	file, err := ParseActionFile(fsys, "site/scroll.bdd")
	require.NoError(t, err)
	require.Len(t, file.Actions, 3)
	down := file.Actions[1].(*Click)
	require.Equal(t, 400, *down.Scroll)
	require.Len(t, down.Asserts, 1)
	require.Equal(t, "SCROLL -100", file.Actions[2].ToFlatString())

	_, err = ParseActionFile(fsys, "site/bad.bdd")
	require.ErrorContains(t, err, "site/bad.bdd:2: SCROLL needs a number of pixels")
}
