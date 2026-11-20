package logview

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestErrorFromAndMeaningReadAnErrorsAttributes(t *testing.T) {
	for _, tc := range []struct {
		attributes string
		from       string
		meaning    string
	}{
		{`{"kind":"client","status":401,"method":"GET","route":"/api/v1/entries","identity":"watch","client":"Nightguard"}`, "watch · GET /api/v1/entries · Nightguard", "token or password is wrong"},
		{`{"kind":"client","status":403,"method":"POST","route":"/api/v1/treatments","identity":"read_only"}`, "read_only · POST /api/v1/treatments", "role does not allow"},
		{`{"kind":"client","status":429,"identity":"pwd_phone"}`, "pwd_phone", "too many requests"},
		{`{"kind":"client","status":400,"method":"POST","route":"/settings/identities","identity":"admin"}`, "admin · POST /settings/identities", "did not accept"},
		{`{"kind":"client","status":404}`, "", "does not exist"},
		{`{"kind":"server","status":500,"method":"GET","route":"/reports"}`, "GET /reports", "Nighttroop failed"},
		{`{"kind":"connector","connector_type":"dexcom"}`, "dexcom connector", "Connectors page"},
		{`{"kind":"site_backup"}`, "", "backup"},
		{`{}`, "", ""},
		{`not json`, "", ""},
	} {
		raw := json.RawMessage(tc.attributes)
		if got := ErrorFrom(context.Background(), raw); got != tc.from {
			t.Errorf("ErrorFrom(%s) = %q, want %q", tc.attributes, got, tc.from)
		}
		if got := ErrorMeaning(context.Background(), raw); (tc.meaning == "") != (got == "") || !strings.Contains(got, tc.meaning) {
			t.Errorf("ErrorMeaning(%s) = %q, want it to contain %q", tc.attributes, got, tc.meaning)
		}
	}
}
