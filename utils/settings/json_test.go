package settings

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// Every setting, as the defaults have each one set, is stored and read back as it was
func TestSettingsJSONRoundTrip(t *testing.T) {
	stored, err := DefaultSettingsMap.JSON()
	require.NoError(t, err)
	again, err := FromJSON(stored, nil, nil).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, string(stored), string(again))
	for _, name := range allSettingNames {
		assert.Contains(t, string(stored), `"`+name+`"`, "a default is set for every setting")
	}
}

// A level with only some settings set stores only those, so the level below shows through for the rest
func TestSettingsJSONLeavesOutWhatIsNotSet(t *testing.T) {
	m := &SettingMap{}
	m.SetTimezone(new("Pacific/Auckland"))
	stored, err := m.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"timezone":"Pacific/Auckland"}`, string(stored))
	read := FromJSON(stored, nil, nil)
	assert.Equal(t, "Pacific/Auckland", *read.GetTimezone())
	assert.Nil(t, read.GetUnits())
}

// A stored row is not trusted: what is not a setting, or not a valid value of one, is left out, and the rest kept
func TestSettingsFromJSONValidates(t *testing.T) {
	read := FromJSON([]byte(`{"timezone":"Pacific/Auckland","units":"stone","range_low_mgdl":"low","not_a_setting":{"a":[1,2,3]},"is_admin":true}`), nil, nil)
	stored, err := read.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"timezone":"Pacific/Auckland"}`, string(stored))

	for name, raw := range map[string]string{"not JSON": `{"timezone":`, "not an object": `["timezone"]`, "empty": ``,
		"too long": `{"timezone":"Pacific/Auckland","pad":"` + strings.Repeat("x", MaxJSONBytes) + `"}`} {
		stored, err := FromJSON([]byte(raw), nil, nil).JSON()
		require.NoError(t, err, name)
		assert.JSONEq(t, `{}`, string(stored), name)
	}
}

// Every setting set, each as long as it may be, fits a row with room to spare, so a row's limit is never what
// refuses a setting
func TestSettingsJSONFitsItsRow(t *testing.T) {
	m := FromJSON(must(DefaultSettingsMap.JSON()), nil, nil)
	devices := make(types.DeviceNames, maxDeviceNames)
	for i := range devices {
		devices[i] = strings.Repeat(string(rune('a'+i)), maxDeviceName)
	}
	m.SetAlarmDeviceBatteryLowDevicesJson(&devices)
	stored, err := m.JSON()
	require.NoError(t, err)
	assert.Less(t, len(stored), MaxJSONBytes/2)
}

func must(raw []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return raw
}

// BenchmarkSettingsFromJSON is reading one level with every setting set, without its query
// Run with go test -run XXX -bench SettingsFromJSON ./utils/settings
func BenchmarkSettingsFromJSON(b *testing.B) {
	stored := must(DefaultSettingsMap.JSON())
	for b.Loop() {
		if FromJSON(stored, nil, nil).GetTimezone() == nil {
			b.Fatal("settings not read")
		}
	}
}

// testdata/released.json has every setting a release has stored, with a value it accepted. A setting renamed or
// removed, or a value no longer accepted, fails here, as a stored row with it would be left out and the person given
// the default with no word. To change one on purpose, read the old name or value where FromJSON reads a row, then
// change the file. A new setting is added to the file with its release.
func TestEveryReleasedSettingIsStillRead(t *testing.T) {
	released, err := os.ReadFile("testdata/released.json")
	require.NoError(t, err)
	read, err := FromJSON(released, nil, nil).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, string(released), string(read), "a released setting is not read as it was stored")
	for _, name := range allSettingNames {
		assert.Contains(t, string(released), `"`+name+`"`, "add the new setting to testdata/released.json")
	}
}

// A save writes only what changed, as a patch of the row: a setting set, one changed, and null for one cleared
func TestChangesAreOnlyWhatChanged(t *testing.T) {
	stored := `{"timezone":"Pacific/Auckland","units":"mg/dL","range_low_mgdl":80}`
	level := FromJSON([]byte(stored), nil, nil)
	patch, names, err := level.Changes(stored)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, patch)
	assert.Empty(t, names)

	level.SetUnits(nil)
	require.NoError(t, level.SetSettingByName("range_low_mgdl", 90, nil, nil))
	require.NoError(t, level.SetSettingByName("alarm_low_enabled", true, nil, nil))
	patch, names, err = level.Changes(stored)
	require.NoError(t, err)
	assert.JSONEq(t, `{"units":null,"range_low_mgdl":90,"alarm_low_enabled":true}`, patch)
	assert.Equal(t, []string{"units", "alarm_low_enabled", "range_low_mgdl"}, names)
}

// A stored value the code no longer accepts is left out of what is read, and is not in a save's patch, so it stays
// in its row
func TestAStoredValueNotValidIsNotChangedByASave(t *testing.T) {
	stored := `{"timezone":"Pacific/Auckland","units":"stone","gone_setting":5}`
	level := FromJSON([]byte(stored), nil, nil)
	assert.Nil(t, level.GetUnits())
	level.SetTimezone(new("Europe/London"))
	patch, names, err := level.Changes(stored)
	require.NoError(t, err)
	assert.JSONEq(t, `{"timezone":"Europe/London"}`, patch)
	assert.Equal(t, []string{"timezone"}, names)
}
