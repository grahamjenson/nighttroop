// Package script is what the .http, .sio and .bdd test files share: reading their lines, with each file they INCLUDE
// in its place, and filling in their {{ }} values
package script

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

// Line is one line of a script that is not blank or a comment, split into its first word and the rest
type Line struct {
	// Path is the file the line is in, which is an INCLUDEd file's for its lines
	Path string
	// From is the files that INCLUDEd the line's file, the outermost first, and none for the file read
	From   []string
	Number int
	// Text is the whole line, trimmed
	Text  string
	First string
	Rest  string
}

// ReadLines reads the lines of the file at path in fsys
func ReadLines(fsys fs.FS, path string) ([]Line, error) {
	contents, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	return SplitLines(contents, path, fsys)
}

// SplitLines is the lines of a script's contents, leaving out blank lines and those starting with # or //. An
// INCLUDE line is the lines of the file it names in fsys, as if written out there, so what follows goes on from the
// included file's last line; with no fsys an INCLUDE is refused.
func SplitLines(contents []byte, path string, fsys fs.FS) ([]Line, error) {
	return splitLines(contents, path, fsys, nil)
}

// splitLines is SplitLines for a file the files from INCLUDEd
func splitLines(contents []byte, path string, fsys fs.FS, from []string) ([]Line, error) {
	var lines []Line
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for number := 1; scanner.Scan(); number++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") || strings.HasPrefix(text, "//") {
			continue
		}
		first, rest, _ := strings.Cut(text, " ")
		rest = strings.TrimSpace(rest)
		if first != "INCLUDE" {
			lines = append(lines, Line{Path: path, From: from, Number: number, Text: text, First: first, Rest: rest})
			continue
		}
		if fsys == nil {
			return nil, fmt.Errorf("%s:%d: INCLUDE %s: no files to include from", path, number, rest)
		}
		contents, err := fs.ReadFile(fsys, rest)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: INCLUDE %s: %w", path, number, rest, err)
		}
		included, err := splitLines(contents, rest, fsys, append(slices.Clone(from), path))
		if err != nil {
			return nil, fmt.Errorf("%s:%d: INCLUDE %s: %w", path, number, rest, err)
		}
		lines = append(lines, included...)
	}
	return lines, scanner.Err()
}
