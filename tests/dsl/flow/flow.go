// Package flow parses and runs a test script of API, Socket.IO and browser steps, as .http, .sio and .bdd files
// are, in any mix: a .flow file can call the API, open a socket, sign in on a page and check what each saw. Each run
// of browser lines is parsed as a .bdd file's, and each run of the others as a .sio file's, so every step and assert
// is written as it is in those files. A browser line starts with a browser action, such as GOTO or CLICK, or is a
// POST of a path on the page's site, with its JSON on the same line; a NAME starts an API or socket step, whose
// POST is of a whole URL, and every other line goes with the step before.
package flow

import (
	"io/fs"
	"slices"
	"strings"

	"nighttroop.t1d.nz/tests/dsl/browser"
	"nighttroop.t1d.nz/tests/dsl/script"
	"nighttroop.t1d.nz/tests/dsl/socketio"
)

// browserActions are the lines that start a browser step
var browserActions = []string{"GOTO", "CLICK", "CLICK_CORNER", "PRESS", "TYPE", "SCROLL", "BACKGROUND", "FOREGROUND",
	"CLICK_BACK", "CLICK_FORWARD", "INPUT", "SAVE_COOKIE", "SET_COOKIE", "WINDOW", "USER_AGENT", "SAVE_TEXT", "CLEAR_COOKIE"}

// Step is one step of a flow: an API or socket step, or a browser action
type Step struct {
	API     *socketio.Step
	Browser browser.Action
}

// Flow is a parsed script
type Flow struct {
	Path  string
	Steps []Step
	// API has every API and socket step, so a socket connection listens for every event the flow waits for
	API *socketio.File
}

// Parse reads the flow at path in fsys, and the files it INCLUDEs
func Parse(fsys fs.FS, path string) (*Flow, error) {
	lines, err := script.ReadLines(fsys, path)
	if err != nil {
		return nil, err
	}
	flow := &Flow{Path: path, API: &socketio.File{Path: path}}
	isBrowser := false
	var run []script.Line
	add := func() error {
		if len(run) == 0 {
			return nil
		}
		if isBrowser {
			actions, err := browser.ParseLines(run)
			for _, action := range actions {
				flow.Steps = append(flow.Steps, Step{Browser: action})
			}
			return err
		}
		file, err := socketio.ParseLines(run, path)
		if err != nil {
			return err
		}
		for _, step := range file.Steps {
			flow.Steps = append(flow.Steps, Step{API: step})
		}
		flow.API.Steps = append(flow.API.Steps, file.Steps...)
		return nil
	}
	for _, line := range lines {
		startsBrowser := slices.Contains(browserActions, line.First) || (line.First == "POST" && strings.HasPrefix(line.Rest, "/"))
		if startsBrowser != isBrowser && (startsBrowser || line.First == "NAME") {
			if err := add(); err != nil {
				return nil, err
			}
			isBrowser, run = startsBrowser, nil
		}
		run = append(run, line)
	}
	if err := add(); err != nil {
		return nil, err
	}
	return flow, nil
}
