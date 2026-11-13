package nsv1

import "testing"

func TestLeftOutSummary(t *testing.T) {
	tests := []struct {
		name    string
		reasons []string
		stored  int
		want    string
	}{
		{"one", []string{"unknown entry type: bad"}, 2, "Stored 2 entries and left out 1: 1 unknown entry type: bad"},
		{"grouped in the order first seen", []string{"device is too long", "glucose out of range", "device is too long"}, 5,
			"Stored 5 entries and left out 3: 2 device is too long; 1 glucose out of range"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left leftOut
			for _, reason := range tt.reasons {
				left.add(reason)
			}
			if got := left.summary(tt.stored); got != tt.want {
				t.Errorf("summary() = %q, want %q", got, tt.want)
			}
		})
	}
}
