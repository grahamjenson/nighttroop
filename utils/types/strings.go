package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type ShortString string

func (s *ShortString) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}
	var str string
	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	default:
		str = fmt.Sprintf("%v", v)
	}

	if len(str) > 64 {
		str = str[:64]
	}

	*s = ShortString(str)
	return nil
}

func (s *ShortString) String() string {
	return string(*s)
}

func (s *ShortString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

type SanitizedString string

func (s *SanitizedString) String() string {
	return string(*s)
}

func (s *SanitizedString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func StripAllTags(input string) string {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return input
	}

	var buf bytes.Buffer
	var f func(*html.Node)

	f = func(n *html.Node) {
		// Only keep text nodes, skip script and style contents
		if n.Type == html.TextNode {
			if n.Parent != nil && (n.Parent.Data == "script" || n.Parent.Data == "style") {
				return
			}
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return buf.String()
}

func (n *SanitizedString) UnmarshalJSON(data []byte) error {
	// We are going to sanitize notes here, no html allowed, trim white space, convert to string
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}
	var str string
	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	default:
		str = fmt.Sprintf("%v", v)
	}

	// Trim white space
	str = strings.TrimSpace(str)
	// Concat to 1024 characters
	if len(str) > 1024 {
		str = str[:1024]
	}
	// Strip all tags
	str = StripAllTags(str)

	*n = SanitizedString(str)
	return nil
}
