package browser

import (
	"bytes"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/mxschmitt/playwright-go"
	"golang.org/x/net/html"
)

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

func renderTemplate(input string, ctx map[string]any) (string, error) {
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"server_url": func() string {
				return ctx["server_url"].(string)
			},
			"site": func() string {
				if v, ok := ctx["site"]; ok && v != nil {
					return ctx["site"].(string)
				}
				return ctx["default_site"].(string)
			},
			// now is the time in a timezone, e.g. {{ now "Pacific/Auckland" "2006-01-02 15:" }}
			"now": func(zone string, layout string) (string, error) {
				location, err := time.LoadLocation(zone)
				if err != nil {
					return "", err
				}
				return time.Now().In(location).Format(layout), nil
			},
			"jwt": func() string {
				if v, ok := ctx["jwt"]; ok && v != nil {
					return ctx["jwt"].(string)
				}
				return ctx["default_jwt"].(string)
			},
		}).
		Parse(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getLocator(expected any, page playwright.Page) playwright.Locator {
	s, ok := expected.(string)
	if !ok {
		return nil
	}

	s = strings.TrimSpace(s)

	switch {
	case strings.HasPrefix(s, "#"):
		// CSS ID: #submit-button
		return page.Locator(s)

	case strings.HasPrefix(s, "."):
		// CSS class: .main-content
		return page.Locator(s)

	case len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`):
		// Text content: "Submit"
		text := strings.Trim(s, `"`)
		return page.GetByText(text)

	default:
		// Normal string value
		return nil
	}

}
