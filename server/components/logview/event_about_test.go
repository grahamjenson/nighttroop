package logview

import (
	"context"
	"encoding/json"
	"testing"
)

func TestEventAboutSaysWhatAnEventsAttributesSayInAFewWords(t *testing.T) {
	for _, tc := range []struct {
		attributes string
		want       string
	}{
		{`{"identity_name":"admin","from":{}}`, "admin"},
		{`{"identity_name":"mum","from":{"city":"Wellington","country":"NZ"}}`, "mum · Wellington, NZ"},
		{`{"identity_name":"mum","from":{"country":"NZ"}}`, "mum · NZ"},
		{`{"identity_name":"watch","by":"admin"}`, "watch · by admin"},
		{`{"identity_name":"admin","by":"admin"}`, "admin"},
		{`{"connector_type":"dexcom","records":12}`, "dexcom · 12 records"},
		{`{"connector_type":"nightscout","from":1700000000000,"to":1700003600000,"records":0}`, "nightscout · 0 records"},
		{`{"by":"admin","from":"mmol","to":"mgdl"}`, "by admin"},
		{`{"reason":"first sync failed"}`, "first sync failed"},
		{`{}`, ""},
		{`not json`, ""},
	} {
		if got := EventAbout(context.Background(), json.RawMessage(tc.attributes)); got != tc.want {
			t.Errorf("EventAbout(%s) = %q, want %q", tc.attributes, got, tc.want)
		}
	}
}
