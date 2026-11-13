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

// A dated page's Next is the rows older than its last row, by its time and ID, however far back, and a page from a
// time shows the way back to the newest rather than which rows from the first it shows
func TestDatedPager(t *testing.T) {
	before := int64(1000)
	tests := []struct {
		name  string
		props Props
		has   []string
		not   []string
	}{
		{"first page", Props{ID: "p", Path: "/data/glucose", Count: 2, Limit: 2, Dated: true, LastAt: 900, LastID: "a b"},
			[]string{`id="p--next"`, `/data/glucose?skip=0&amp;limit=2&amp;before=900&amp;before_id=a+b`, `id="p--date"`, `data-pager-skip`, `>Older<`}, []string{`id="p--newest"`, `to {to}" hidden`}},
		{"past the most skipped", Props{ID: "p", Path: "/data/glucose", Count: 2, Skip: apiquery.MaxQuerySkip, Limit: 2, Dated: true, LastAt: 900, LastID: "a"},
			[]string{`id="p--next"`}, nil},
		{"from a time", Props{ID: "p", Path: "/data/glucose", Count: 2, Limit: 2, Dated: true, Before: &before, LastAt: 900, LastID: "a"},
			[]string{`id="p--newest"`, `/data/glucose?skip=0&amp;limit=2"`, `id="p--next"`, `value="1970-01-01"`, `to {to}" hidden`}, nil},
		{"last page from a time", Props{ID: "p", Path: "/data/glucose", Count: 1, Limit: 2, Dated: true, Before: &before},
			[]string{`id="p--newest"`}, []string{`id="p--next"`}},
		{"by skip", Props{ID: "p", Path: "/presets/food", Count: 2, Limit: 2},
			[]string{`/presets/food?skip=2&amp;limit=2`, `>Next<`}, []string{`id="p--date"`, `before=`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var html bytes.Buffer
			if err := Pager(tt.props).Render(context.Background(), &html); err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.has {
				if !strings.Contains(html.String(), want) {
					t.Errorf("missing %s in %s", want, html.String())
				}
			}
			for _, unwanted := range tt.not {
				if strings.Contains(html.String(), unwanted) {
					t.Errorf("unexpected %s in %s", unwanted, html.String())
				}
			}
		})
	}
}
