package siteapi

import (
	"net/url"
	"testing"
)

func TestFormToMap(t *testing.T) {
	type model struct {
		Enabled bool    `json:"enabled"`
		Off     bool    `json:"off"`
		Count   int     `json:"count"`
		Name    *string `json:"name"`
	}
	doc := FormToMap[model](url.Values{"enabled": {"on"}, "count": {"3"}, "name": {"true"}, "a.b": {"c"}})
	if doc["enabled"] != true || doc["off"] != false || doc["count"] != 3.0 || doc["name"] != "true" {
		t.Fatalf("unexpected conversion: %v", doc)
	}
	if nested, _ := doc["a"].(map[string]any); nested["b"] != "c" {
		t.Fatalf("dotted name did not nest: %v", doc)
	}
}
