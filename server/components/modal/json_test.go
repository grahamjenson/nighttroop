package modal

import "testing"

func TestIndentJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"object", `{"sgv":120,"type":"sgv"}`, "{\n  \"sgv\": 120,\n  \"type\": \"sgv\"\n}"},
		{"nested", `{"pump":{"reservoir":12}}`, "{\n  \"pump\": {\n    \"reservoir\": 12\n  }\n}"},
		{"not JSON is left as it is", `not json`, `not json`},
		{"empty", ``, ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indentJSON(tt.raw); got != tt.want {
				t.Errorf("indentJSON(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
