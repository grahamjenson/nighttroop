package addmodal

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEdit(t *testing.T) {
	type record struct {
		Dose       float64  `json:"dose"`
		OccurredAt int64    `json:"occurred_at"`
		SplitNow   *float64 `json:"split_now"`
		Multiplier *float64 `json:"multiplier"`
		Notes      *string  `json:"notes"`
		Kind       string   `json:"kind"`
		Reading    float64  `json:"reading"`
	}
	fields := []Field{
		{Name: "dose"}, {Name: "occurred_at", Kind: When}, {Name: "split_now", Kind: Percent},
		{Name: "multiplier", Kind: Percent, Value: "100"}, {Name: "notes", Kind: Text}, {Name: "kind", Kind: Select},
		{Name: "reading"},
	}
	notes := "Lunch"
	r := &record{Dose: 4.5, OccurredAt: 1700000000000, SplitNow: new(0.7), Notes: &notes, Kind: "normal", Reading: 5.5494}
	edit := Edit(fields, r, map[string]string{"reading": "5.5"}, "kind")

	values := map[string]string{}
	for _, f := range edit {
		values[f.Name] = f.Value
	}
	require.Equal(t, map[string]string{
		"dose": "4.5",
		// A fraction is shown as the percent it is typed in as
		"split_now": "70",
		// A field the record has no value for starts as in the add modal
		"multiplier": "100",
		"notes":      "Lunch",
		// A value given starts at that rather than the record's
		"reading": "5.5",
		// Its time and the fixed fields are left out, as an update does not change them
	}, values)
	require.Len(t, edit, 5)
	// The add modal's fields are left as they were
	require.Equal(t, "", fields[0].Value)
}

func TestDetail(t *testing.T) {
	edit := Props{Fields: []Field{{Name: "notes", Label: "Notes"}}}
	tests := []struct {
		name, field, label, value, want string
	}{
		{"a field the edit form has is left out, as the form shows it", "notes", "Notes", "Lunch", ""},
		{"a field it does not have is listed", "dia", "DIA", "4", `<dt class="text-gray-500">DIA</dt><dd class="break-all">4</dd>`},
		{"a field with no value is left out", "dia", "DIA", "-", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, Detail(edit, tt.field, tt.label, tt.value).Render(context.Background(), &buf))
			require.Equal(t, tt.want, buf.String())
		})
	}
}
