package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A profile switch's other attributes are NULL with none set, and otherwise a JSON object of only those set, which
// reads back, a timeshift and percentage of 0 kept
func TestProfileSwitchOther(t *testing.T) {
	name, customized := "Default", "Default (110%)"
	timeshift, percentage, duration, end := int64(0), int64(110), int64(3600000), int64(1767225600000)
	zero := int64(0)
	for _, test := range []struct {
		name  string
		other types.ProfileSwitchOther
		value any
	}{
		{"none set is NULL", types.ProfileSwitchOther{}, nil},
		{"only those set", types.ProfileSwitchOther{OriginalProfileName: &name}, `{"original_profile_name":"Default"}`},
		{"a switch it was made from", types.ProfileSwitchOther{OriginalProfileName: &name, OriginalCustomizedName: &customized,
			OriginalTimeshiftMs: &timeshift, OriginalPercentage: &percentage, OriginalDurationMs: &duration, OriginalEnd: &end},
			`{"original_profile_name":"Default","original_customized_name":"Default (110%)","original_timeshift_ms":0,` +
				`"original_percentage":110,"original_duration_ms":3600000,"original_end":1767225600000}`},
		{"a percentage of 0", types.ProfileSwitchOther{OriginalPercentage: &zero}, `{"original_percentage":0}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.ProfileSwitchOther
			require.NoError(t, read.Scan(value))
			require.Equal(t, test.other, read)

			// An update sends those set as JSON, and nothing for none
			if json := types.OtherJSON(test.other); test.value == nil {
				require.Nil(t, json)
			} else {
				require.Equal(t, test.value, *json)
			}
		})
	}
}
