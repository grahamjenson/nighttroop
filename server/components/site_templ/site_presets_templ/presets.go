package site_presets_templ

import (
	"fmt"

	"nighttroop.t1d.nz/server/components/modal"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
)

// formatDate shows times as the data pages do
var formatDate = site_data_templ.FormatDate

// orDash shows an optional value, or a dash when it is not set
func orDash[T any](value *T) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprint(*value)
}

func orEmpty[T any](value *T) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(*value)
}

var presetDetail = modal.DetailIfSet
