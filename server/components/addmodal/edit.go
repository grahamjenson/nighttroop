package addmodal

import (
	"math"
	"slices"
	"strconv"

	"nighttroop.t1d.nz/utils"
)

// Edit is an add modal's fields to change a record with in its details modal, with the same labels and help, each
// starting at the record's value as Fill starts it. It leaves out the record's time and the fields named in fixed,
// which an update does not change.
func Edit[T any](fields []Field, record *T, values map[string]string, fixed ...string) []Field {
	edit := slices.DeleteFunc(slices.Clone(fields), func(f Field) bool { return f.Kind == When || slices.Contains(fixed, f.Name) })
	return Fill(edit, record, values)
}

// Fill is fields each starting at the record's value, by JSON name, such as an add modal's at a preset's, but its time,
// which a new record takes from when it is added. A field the record has no value for starts as it was, and one in
// values, such as a glucose reading rounded as it is shown, at that.
func Fill[T any](fields []Field, record *T, values map[string]string) []Field {
	recorded, _ := utils.ToMap(record)
	filled := []Field{}
	for _, f := range fields {
		if value, ok := values[f.Name]; ok {
			f.Value = value
		} else if value := formatValue(f.Kind, recorded[f.Name]); value != "" && f.Kind != When {
			f.Value = value
		}
		filled = append(filled, f)
	}
	return filled
}

// formatValue is a record's JSON value as its input takes it, a Percent's fraction as a percent, e.g. 1.2 as 120
func formatValue(kind Kind, value any) string {
	switch v := value.(type) {
	case float64:
		if kind == Percent {
			v = math.Round(v*10000) / 100
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	}
	return ""
}
