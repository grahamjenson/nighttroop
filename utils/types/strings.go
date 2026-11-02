package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type JSONString string

func (s *JSONString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("expected JSON string, got %s", data)
	}

	if len(str) > 2048 {
		str = str[:2048]
	}

	*s = JSONString(str)
	return nil
}

func (s *JSONString) String() string {
	return string(*s)
}

type LongJSONString string

func (s *LongJSONString) String() string {
	return string(*s)
}

func (s *LongJSONString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

func (n *LongJSONString) UnmarshalJSON(data []byte) error {
	// We are going to sanitize notes here, no html allowed, trim white space, convert to string
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}

	// Test that this is valid json string
	var testJSON any
	err := json.Unmarshal([]byte(v), &testJSON)
	if err != nil {
		return fmt.Errorf("invalid JSON string, got %s", data)
	}

	// Concat to 1024 characters
	if len(v) > 4024 {
		v = v[:4024]
	}

	*n = LongJSONString(v)
	return nil
}

type IdentifierString string

func (i *IdentifierString) String() string {
	if i == nil {
		return ""
	}
	return string(*i)
}

func (i *IdentifierString) Stringp() *string {
	if i == nil {
		return nil
	}
	str := string(*i)
	return &str
}

func (i *IdentifierString) UnmarshalJSON(data []byte) error {
	// limit in possible characters and length, convert everything to a string
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected identifier string, got %s", data)
	}

	str := ""

	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		str = v.String()
	default:
		str = fmt.Sprintf("%v", v)
	}

	// 36 fits a UUID, the same limit as identifiers in the URL and the id columns. A longer one, like a Loop dose's
	// syncIdentifier, is hashed rather than cut, as two that start the same, Omnipod's temp basals of the same rate,
	// would otherwise be the same identifier
	if len(str) > 36 {
		sum := sha256.Sum256([]byte(str))
		str = hex.EncodeToString(sum[:])[:36]
	}

	if len(str) == 0 {
		return fmt.Errorf("identifier string cannot be empty")
	}

	*i = IdentifierString(str)
	return nil
}
