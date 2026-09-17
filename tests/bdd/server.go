package bdd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/testutils"
)

func NewRouter(t *testing.T) http.Handler {
	config, controlDB, _, err := testutils.GetConfigControlDBAndSiteDBForTest()

	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)
	return router
}

func PrettifyHTML(input string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var buf bytes.Buffer
	indent := 0

	// Helper to write tabs
	writeIndent := func(level int) {
		for i := 0; i < level; i++ {
			buf.WriteString("\t")
		}
	}

	// Self-closing tags that shouldn't increase nesting/indentation
	selfClosingTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return input
		}

		token := tokenizer.Token()

		switch tokenType {
		case html.StartTagToken:
			// Print current line break and indent before writing the tag
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			writeIndent(indent)
			buf.WriteString(token.String())

			// Only indent next line if it's not a self-closing tag
			if !selfClosingTags[token.Data] {
				indent++
			}

		case html.EndTagToken:
			indent--
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			writeIndent(indent)
			buf.WriteString(token.String())

		case html.SelfClosingTagToken:
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			writeIndent(indent)
			buf.WriteString(token.String())

		case html.TextToken:
			// Clean up extra whitespace/newlines in text nodes
			text := strings.TrimSpace(token.Data)
			if text != "" {
				if buf.Len() > 0 {
					buf.WriteString("\n")
				}
				writeIndent(indent)
				buf.WriteString(text)
			}

		case html.CommentToken, html.DoctypeToken:
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			writeIndent(indent)
			buf.WriteString(token.String())
		}
	}

	return buf.String()
}
