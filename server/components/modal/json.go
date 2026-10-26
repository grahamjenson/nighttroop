package modal

import (
	"bytes"
	"encoding/json"
)

// indentJSON indents JSON to read, leaving anything that is not JSON as it is
func indentJSON(raw string) string {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(raw), "", "  "); err != nil {
		return raw
	}
	return out.String()
}
