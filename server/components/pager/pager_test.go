package pager

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"nighttroop.t1d.nz/utils/apiutils/apiquery"
)

func TestPageURL(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"path", "/data/insulin", "/data/insulin?skip=50&limit=25"},
		{"path with a query", "/data/device-status?device=pump", "/data/device-status?device=pump&skip=50&limit=25"},
		{"empty order", "/sites?order=", "/sites?order=&skip=50&limit=25"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PageURL(tt.path, 50, 25); got != tt.want {
				t.Errorf("PageURL(%q, 50, 25) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// A full page has Next, unless its next page would start past the furthest a search skips, which would show it again
func TestPagerNextStopsAtMaxQuerySkip(t *testing.T) {
	tests := []struct {
		skip int64
		next bool
	}{
		{0, true},
		{apiquery.MaxQuerySkip - 100, true},
		{apiquery.MaxQuerySkip - 50, true},
		{apiquery.MaxQuerySkip, false},
	}
	for _, tt := range tests {
		var html bytes.Buffer
		if err := Pager(Props{ID: "p", Path: "/data/glucose", Count: 50, Skip: tt.skip, Limit: 50}).Render(context.Background(), &html); err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(html.String(), `id="p--next"`); got != tt.next {
			t.Errorf("skip %d: Next shown %v, want %v", tt.skip, got, tt.next)
		}
	}
}
