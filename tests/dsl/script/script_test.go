package script

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

// Blank lines and comments are left out, and an INCLUDE is the included file's lines, each saying where it is from
func TestReadLines(t *testing.T) {
	fsys := fstest.MapFS{
		"main.http":         {Data: []byte("# a comment\nNAME first\n\n  GET http://{{ site }}/x  \nINCLUDE steps/login.steps\n// another\nASSERT_STATUS 200\n")},
		"steps/login.steps": {Data: []byte("NAME login\nPOST http://{{ site }}/login\n")},
		"bad.http":          {Data: []byte("NAME first\nINCLUDE steps/missing.steps\n")},
	}
	lines, err := ReadLines(fsys, "main.http")
	require.NoError(t, err)
	require.Equal(t, []Line{
		{Path: "main.http", Number: 2, Text: "NAME first", First: "NAME", Rest: "first"},
		{Path: "main.http", Number: 4, Text: "GET http://{{ site }}/x", First: "GET", Rest: "http://{{ site }}/x"},
		{Path: "steps/login.steps", From: []string{"main.http"}, Number: 1, Text: "NAME login", First: "NAME", Rest: "login"},
		{Path: "steps/login.steps", From: []string{"main.http"}, Number: 2, Text: "POST http://{{ site }}/login", First: "POST", Rest: "http://{{ site }}/login"},
		{Path: "main.http", Number: 7, Text: "ASSERT_STATUS 200", First: "ASSERT_STATUS", Rest: "200"},
	}, lines)

	_, err = ReadLines(fsys, "bad.http")
	require.ErrorContains(t, err, "bad.http:2: INCLUDE steps/missing.steps")
	_, err = SplitLines([]byte("INCLUDE steps/login.steps"), "inline.sio", nil)
	require.ErrorContains(t, err, "inline.sio:1: INCLUDE steps/login.steps: no files to include from")
}

// One set of helpers fills in every kind of script's values
func TestRender(t *testing.T) {
	ctx := map[string]any{
		"server_url":   "http://127.0.0.1:1234",
		"default_site": "abcdefgh.ns.localhost",
		"created":      map[string]any{"entries": []any{map[string]any{"sgv": 120.0}}},
	}
	for input, want := range map[string]string{
		"{{ server_url }}/login":            "http://127.0.0.1:1234/login",
		"http://{{ site }}/api":             "http://abcdefgh.ns.localhost/api",
		`{{ dig "created.entries.0.sgv" }}`: "120",
		`{{ dig "created.entries.1.sgv" }}`: "no_value",
		`{{ replace ".ns." "." site }}`:     "abcdefgh.localhost",
		`{{ sha1 "token" }}`:                "ee977806d7286510da8b9a7492ba58e2484c0ecc",
	} {
		got, err := Render(input, ctx)
		require.NoError(t, err, input)
		require.Equal(t, want, got, input)
	}
	_, err := Render("{{ .missing }}", ctx)
	require.Error(t, err)
}
