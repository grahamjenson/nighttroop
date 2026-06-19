package collections

import (
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func CreateRecordSlice[T any](in []T, fn func(T) nsutils.Record) []nsutils.Record {
	out := make([]nsutils.Record, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}
