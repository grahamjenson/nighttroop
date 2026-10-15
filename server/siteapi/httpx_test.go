package siteapi

import (
	"net/url"
	"testing"
)

func TestCoerceFormValues(t *testing.T) {
	type model struct {
		Enabled bool    `json:"enabled"`
		Off     bool    `json:"off"`
		Count   int     `json:"count"`
		Name    *string `json:"name"`
	}
	doc := FormToMap(url.Values{"enabled": {"on"}, "count": {"3"}, "name": {"true"}})
	coerceFormValues(doc, new(model))
	if doc["enabled"] != true || doc["off"] != false || doc["count"] != 3.0 || doc["name"] != "true" {
		t.Fatalf("unexpected coercion: %v", doc)
	}
}
