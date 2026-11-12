package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

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
	// We are going to sanitize notes here, no html allowed, convert to string
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

	// Concat to 1024 characters
	if len(str) > 1024 {
		str = str[:1024]
	}
	// Strip all tags between the white space around the text, which parsing it as HTML would drop
	start := len(str) - len(strings.TrimLeftFunc(str, unicode.IsSpace))
	end := start + len(strings.TrimRightFunc(str[start:], unicode.IsSpace))
	str = str[:start] + StripAllTags(str[start:end]) + str[end:]
	// Only white space is none; other white space around the text is kept as sent, as Nightscout keeps it
	if strings.TrimSpace(str) == "" {
		str = ""
	}

	*n = SanitizedString(str)
	return nil
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

// MaxJSONTextLength is the longest JSON text a JSONTextString keeps, four times AAPS's Bolus Wizard calculation
const MaxJSONTextLength = 4096

// JSONTextString is JSON text sent as a string, as AAPS sends its Bolus Wizard's calculation, kept exactly as sent. One
// that is not JSON or is longer than MaxJSONTextLength is empty, which is not stored, rather than refuse the upload
type JSONTextString string

func (s *JSONTextString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

func (s *JSONTextString) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected a JSON string, got %s", data)
	}
	if len(v) > MaxJSONTextLength || !json.Valid([]byte(v)) {
		v = ""
	}
	*s = JSONTextString(v)
	return nil
}

// MaxDeviceTokenLength is the longest device token kept: Apple's APNs tokens are of no fixed length, 64 hex characters
// once and 160 from Trio now
const MaxDeviceTokenLength = 256

// DeviceTokenString is the token of an app on a device, as Loop and Trio send their APNs token, kept whole. One longer
// than MaxDeviceTokenLength is empty, which is not registered for pushes, rather than cut or refuse the upload
type DeviceTokenString string

func (s *DeviceTokenString) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected a device token string, got %s", data)
	}
	if len(v) > MaxDeviceTokenLength {
		v = ""
	}
	*s = DeviceTokenString(v)
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

	if len(str) == 0 {
		return fmt.Errorf("identifier string cannot be empty")
	}

	*i = *ToIdentifierString(str)
	return nil
}

// IdentifierFromJSON is the identifier in data, or nil when there is none or it is not one
func IdentifierFromJSON(data json.RawMessage) *IdentifierString {
	var id IdentifierString
	if len(data) == 0 || string(data) == "null" || json.Unmarshal(data, &id) != nil {
		return nil
	}
	return &id
}

// ToIdentifierString is str as an identifier, or nil when it is empty. 36 fits a UUID, the same limit as identifiers in
// the URL and the id columns. A longer one, like a Loop dose's syncIdentifier, is hashed rather than cut, as two that
// start the same, Omnipod's temp basals of the same rate, would otherwise be the same identifier
func ToIdentifierString(str string) *IdentifierString {
	if len(str) == 0 {
		return nil
	}
	if len(str) > 36 {
		sum := sha256.Sum256([]byte(str))
		str = hex.EncodeToString(sum[:])[:36]
	}
	return new(IdentifierString(str))
}
