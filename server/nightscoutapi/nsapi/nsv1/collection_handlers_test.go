package nsv1

import (
	"testing"

	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

func TestLeftOutSummary(t *testing.T) {
	tests := []struct {
		name       string
		collection nsutils.Collection
		reasons    []string
		stored     int
		want       string
	}{
		{"one", nsutils.Entries, []string{"unknown entry type: bad"}, 2, "Stored 2 entries and left out 1: 1 unknown entry type: bad"},
		{"grouped in the order first seen", nsutils.Entries, []string{"device is too long", "glucose out of range", "device is too long"}, 5,
			"Stored 5 entries and left out 3: 2 device is too long; 1 glucose out of range"},
		{"named by its collection", nsutils.Treatments, []string{"unknown eventType"}, 1, "Stored 1 treatments and left out 1: 1 unknown eventType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left leftOut
			for _, reason := range tt.reasons {
				left.add(reason)
			}
			if got := left.summary(tt.collection, tt.stored); got != tt.want {
				t.Errorf("summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLeftOutLeaves checks which problems leave a document out of a batch and which stop it
func TestLeftOutLeaves(t *testing.T) {
	bad := nsutils.BadDataBodyError
	tests := []struct {
		name string
		errs []*nsutils.HTTPError
		want []bool
		bad  int
	}{
		{"stored", []*nsutils.HTTPError{nil}, []bool{false}, 0},
		{"a problem with the document", []*nsutils.HTTPError{bad, nsutils.BadDateFieldError}, []bool{true, true}, 2},
		{"a database CHECK", []*nsutils.HTTPError{nsutils.BadDataBodyErrorWithHint("dose is out of range")}, []bool{true}, 1},
		{"readings out of range are not counted", []*nsutils.HTTPError{nsutils.BadGlucoseError, nsutils.BadGlucoseError}, []bool{true, true}, 0},
		{"credentials stop the batch", []*nsutils.HTTPError{nsutils.ErrMissingCredentials}, []bool{false}, 0},
		{"permissions stop the batch", []*nsutils.HTTPError{nsutils.ErrRecordReadOnly}, []bool{false}, 0},
		{"rate limits stop the batch", []*nsutils.HTTPError{nsutils.ErrTooManyBatchWrites}, []bool{false}, 0},
		{"server errors stop the batch", []*nsutils.HTTPError{nsutils.ErrInternal}, []bool{false}, 0},
		{"an 11th problem stops the batch", []*nsutils.HTTPError{bad, bad, bad, bad, bad, bad, bad, bad, bad, bad, nsutils.BadGlucoseError, bad},
			[]bool{true, true, true, true, true, true, true, true, true, true, true, false}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left leftOut
			for i, err := range tt.errs {
				if got := left.leaves(err); got != tt.want[i] {
					t.Errorf("leaves(%v) #%d = %v, want %v", err, i, got, tt.want[i])
				}
			}
			if left.bad != tt.bad {
				t.Errorf("bad = %d, want %d", left.bad, tt.bad)
			}
		})
	}
}
