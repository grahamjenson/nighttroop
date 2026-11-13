package nsquery

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// Every collection's search keeps its device filter, so a search by device never gives every device's records
func TestSearchesKeepTheDevice(t *testing.T) {
	for _, filter := range []string{"find[device]", "find[device][$eq]", "device", "device$eq"} {
		values, err := ParseQueryValues(url.Values{filter: {"A"}})
		require.NoError(t, err)
		require.False(t, values.HasUnknownFilter(), filter)
		searchesKeepTheDevice(t, filter, QueryValueToQuery(values))
	}
}

func searchesKeepTheDevice(t *testing.T, filter string, query Query) {
	searches := map[string]any{
		"activity": ActivitySearch(query), "devicestatus": DeviceReportSearch(query), "treatments": EventGroupSearch(query),
		"food intake": FoodIntakeSearch(query), "food": FoodPresetSearch(query), "entries": GlucoseSearch(query),
		"profile": ProfileStoreSearch(query),
	}
	for name, params := range searches {
		t.Run(filter+" "+name, func(t *testing.T) {
			device := reflect.ValueOf(params).FieldByName("DeviceEq")
			require.True(t, device.IsValid(), "the search has no DeviceEq")
			require.False(t, device.IsNil(), "the search dropped the device")
			require.Equal(t, "A", device.Elem().String())
		})
	}
}
