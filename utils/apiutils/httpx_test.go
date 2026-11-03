package apiutils_test

import (
	"net/url"
	"reflect"
	"testing"
	"time"

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
	doc := apiutils.FormToMap[model](url.Values{"enabled": {"on"}, "count": {"3"}, "name": {"true"}, "rate": {"0"}, "a.b": {"c"}}, time.UTC)
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
	doc := apiutils.FormToMap[model](url.Values{"minutes": {""}, "rate": {"1.5"}, "notes": {""}}, time.UTC)
	if _, sent := doc["minutes"]; sent {
		t.Fatalf("blank number was sent: %v", doc)
	}
	if doc["rate"] != 1.5 || doc["notes"] != "" {
		t.Fatalf("unexpected conversion: %v", doc)
	}
}

// A date and time input is the milliseconds of that time in the timezone times are shown in
func TestFormToMapDateTime(t *testing.T) {
	type model struct {
		OccurredAt int64 `json:"occurred_at"`
	}
	auckland, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		t.Fatal(err)
	}
	doc := apiutils.FormToMap[model](url.Values{"occurred_at": {"2026-10-31T09:30"}}, auckland)
	if want := time.Date(2026, 10, 31, 9, 30, 0, 0, auckland).UnixMilli(); doc["occurred_at"] != want {
		t.Fatalf("got %v, want %v", doc["occurred_at"], want)
	}
}

// A list is every value sent, each split at commas and spaces, as checkboxes ticked or a text of them; one not sent is
// left out
func TestFormToMapList(t *testing.T) {
	type model struct {
		Apps      []string `json:"apps"`
		BundleIDs []string `json:"bundle_ids"`
		Tags      []string `json:"tags"`
	}
	doc := apiutils.FormToMap[model](url.Values{"apps": {"loop", "trio"}, "bundle_ids": {"a.b, c.d  e.f,"}}, time.UTC)
	if !reflect.DeepEqual(doc["apps"], []string{"loop", "trio"}) || !reflect.DeepEqual(doc["bundle_ids"], []string{"a.b", "c.d", "e.f"}) {
		t.Fatalf("unexpected lists: %v", doc)
	}
	if _, sent := doc["tags"]; sent {
		t.Fatalf("a list not sent was set: %v", doc)
	}
}
