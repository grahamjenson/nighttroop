package controlmodels

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSiteInactive(t *testing.T) {
	now := time.Now()
	daysAgo := func(days int) *int64 {
		ms := now.AddDate(0, 0, -days).UnixMilli()
		return &ms
	}
	tests := []struct {
		name      string
		createdAt *int64
		lastSeen  *int64
		lastWrite *int64
		inactive  bool
	}{
		{"seen and written today", daysAgo(30), daysAgo(0), daysAgo(0), false},
		{"seen today, written 8 days ago", daysAgo(30), daysAgo(0), daysAgo(8), true},
		{"written today, seen 8 days ago", daysAgo(30), daysAgo(8), daysAgo(0), true},
		{"made today, never seen or written", daysAgo(0), nil, nil, false},
		{"made 8 days ago, never seen or written", daysAgo(8), nil, nil, true},
		{"seen and written 6 days ago", daysAgo(30), daysAgo(6), daysAgo(6), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			site := &Site{CreatedAt: *test.createdAt, LastSeenAt: test.lastSeen, LastWriteAt: test.lastWrite}
			require.Equal(t, test.inactive, site.Inactive(7, now))
		})
	}
}
