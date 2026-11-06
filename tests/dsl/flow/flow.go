// Package flow parses and runs a test script of API, Socket.IO and browser steps, as .http, .sio and .bdd files
// are, in any mix: a .flow file can call the API, open a socket or stream, sign in on a page and check what each saw.
// A line HTTP or BROWSE switches the lines after it to HTTP steps, which are written as in a .sio file (an HTTP
// request, CONNECT, EMIT, WAIT, STREAM and the rest), or browser steps, written as in a .bdd file (GOTO, CLICK and
// the rest), each with the asserts after it. A .bdd file starts with browser steps and any other with HTTP ones; a
// .steps file goes on as the file that INCLUDEs it was, and a switch in an included file lasts until its end.
package flow

import (
	"io/fs"
	"path/filepath"
	"slices"

	"nighttroop.t1d.nz/tests/dsl/browser"
	"nighttroop.t1d.nz/tests/dsl/script"
	"nighttroop.t1d.nz/tests/dsl/socketio"
)

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
	// files is a frame for each file the line being read is in, the outermost first: its path, its last line and
	// whether its lines, from its start or its last HTTP or BROWSE, are browser steps
	type file struct {
		path     string
		number   int
		browsing bool
	}
	var files []file
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
		// Leave the files the line is not in, or that it starts again, as a file INCLUDEd twice does
		chain := append(slices.Clone(line.From), line.Path)
		kept := 0
		for kept < len(files) && kept < len(chain) && files[kept].path == chain[kept] &&
			(kept < len(chain)-1 || line.Number > files[kept].number) {
			kept++
		}
		files = files[:kept]
		for _, path := range chain[kept:] {
			// A .bdd file starts browsing, a .steps file as the file that includes it is, and any other with HTTP
			browsing := filepath.Ext(path) == ".bdd"
			if filepath.Ext(path) == ".steps" && len(files) > 0 {
				browsing = files[len(files)-1].browsing
			}
			files = append(files, file{path: path, browsing: browsing})
		}
		current := &files[len(files)-1]
		current.number = line.Number
		if line.First == "HTTP" || line.First == "BROWSE" {
			current.browsing = line.First == "BROWSE"
		}
		if current.browsing != isBrowser {
			if err := add(); err != nil {
				return nil, err
			}
			isBrowser, run = current.browsing, nil
		}
		if line.First != "HTTP" && line.First != "BROWSE" {
			run = append(run, line)
		}
	}
	if err := add(); err != nil {
		return nil, err
	}
	return flow, nil
}
