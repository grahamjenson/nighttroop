package apiutils_test

import (
	"net/url"
	"testing"

	"nighttroop.t1d.nz/utils/apiutils"
)

func TestFormToMap(t *testing.T) {
	type model struct {
		Enabled bool     `json:"enabled"`
		Off     bool     `json:"off"`
		Count   int      `json:"count"`
		Name    *string  `json:"name"`
		Rate    *float64 `json:"rate"`
	}
	doc := apiutils.FormToMap[model](url.Values{"enabled": {"on"}, "count": {"3"}, "name": {"true"}, "rate": {"0"}, "a.b": {"c"}})
	if doc["enabled"] != true || doc["off"] != false || doc["count"] != 3.0 || doc["name"] != "true" || doc["rate"] != 0.0 {
		t.Fatalf("unexpected conversion: %v", doc)
	}
	if nested, _ := doc["a"].(map[string]any); nested["b"] != "c" {
		t.Fatalf("dotted name did not nest: %v", doc)
	}
}

// A number input left blank is not sent, so a JSON decode of it does not fail and an update leaves it as it is
func TestFormToMapBlankNumber(t *testing.T) {
	type model struct {
		Minutes *int64   `json:"minutes"`
		Rate    *float64 `json:"rate"`
		Notes   *string  `json:"notes"`
	}
	doc := apiutils.FormToMap[model](url.Values{"minutes": {""}, "rate": {"1.5"}, "notes": {""}})
	if _, sent := doc["minutes"]; sent {
		t.Fatalf("blank number was sent: %v", doc)
	}
	if doc["rate"] != 1.5 || doc["notes"] != "" {
		t.Fatalf("unexpected conversion: %v", doc)
	}
}
