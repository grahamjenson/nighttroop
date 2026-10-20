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
