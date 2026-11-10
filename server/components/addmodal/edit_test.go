package addmodal

import (
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
