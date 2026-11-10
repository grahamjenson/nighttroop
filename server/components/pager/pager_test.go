package pager

import "testing"

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
