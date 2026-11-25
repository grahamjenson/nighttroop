package settings

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// values_schema.txt is every field of Values: its place in the JSON, its type and its setting. It only gains lines.
//
//go:embed values_schema.txt
var valuesSchema string

// values_v1.json is a row as version 1 stored it, with every setting of that version set
//
//go:embed values_v1.json
var valuesV1 string

// schema is each field of Values as a line: its JSON path, its Go type and the setting kept in it, in the order
// they are declared
func schema() []string {
	v := &Values{}
	settingAt := map[any]string{}
	for _, f := range v.fields() {
		settingAt[f.at] = f.identifier
	}
	return schemaOf(reflect.ValueOf(v).Elem(), "", settingAt)
}

func schemaOf(v reflect.Value, path string, settingAt map[any]string) []string {
	var lines []string
	for i := range v.NumField() {
		field := v.Type().Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		at := strings.TrimPrefix(path+"."+name, ".")
		if field.Type.Kind() == reflect.Struct {
			lines = append(lines, schemaOf(v.Field(i), at, settingAt)...)
			continue
		}
		lines = append(lines, strings.TrimSpace(at+" "+field.Type.String()+" "+settingAt[v.Field(i).Addr().Interface()]))
	}
	return lines
}

// Rows of every release are read by the JSON names of Values, so a field is never renamed, removed, moved, given
// another type or given to another setting: its row would be read as not set, and the person given the default with
// no word. A new field is fine, and is added to values_schema.txt in the same change. Never change or delete a line
// of that file to make this pass.
func TestValuesOnlyGrow(t *testing.T) {
	now := schema()
	released := strings.Split(strings.TrimSpace(valuesSchema), "\n")
	for _, line := range released {
		assert.Contains(t, now, line, "a released field of Values was renamed, removed, moved or retyped: put it back as it was")
	}
	for _, line := range now {
		assert.Contains(t, released, line, "a new field of Values: add this line to the end of values_schema.txt")
	}
	assert.Equal(t, len(released), len(slices.Compact(slices.Sorted(slices.Values(released)))), "a line is in values_schema.txt twice")
}

// A row a release stored is read whole by every later one: each setting it has is read with the value it had, and
// written back the same. To change what a stored value means, add a version and read version 1 as before; never
// change values_v1.json.
func TestARowOfVersion1IsStillRead(t *testing.T) {
	level := LevelFromJSON(valuesV1, "site", 0)
	again, err := level.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, valuesV1, again)

	// And is the values it says, as the struct it is read into
	v := FromLevels(level, nil).Values
	var stored Values
	require.NoError(t, json.Unmarshal([]byte(valuesV1), &stored))
	assert.Equal(t, stored, v)
	assert.Equal(t, "Pacific/Auckland", v.Timezone)
	assert.Equal(t, int64(80), v.Range.LowMgdl)
	assert.Equal(t, int64(10), v.Alarms.Low.Minutes)
	assert.Equal(t, int64(20), v.Alarms.FastRise.Minutes)
	assert.Equal(t, types.DeviceNames{"openaps://pixel", "SugarPixel"}, v.Alarms.DeviceBatteryLow.Devices)
}

// Every setting has one place in Values, and every place is a setting, so none is lost when saved, and with no
// level every one is at its default
func TestEverySettingHasItsField(t *testing.T) {
	assert.Len(t, schema(), len(names)+1, "a field of Values that is no setting's, but for version")
	assert.Len(t, slices.Compact(slices.Sorted(slices.Values(names))), len(names), "a setting is in two fields")
	for _, line := range schema()[1:] {
		assert.Len(t, strings.Fields(line), 3, "a field with no setting: "+line)
	}

	defaults := FromLevels(nil, nil)
	assert.Equal(t, DefaultValues(), defaults.Values)
	assert.Equal(t, Version, defaults.Version)
	assert.Equal(t, "UTC", defaults.Timezone)
	assert.Equal(t, int64(72), defaults.Range.LowMgdl)
	assert.False(t, defaults.Alarms.Low.Enabled)
	for _, setting := range defaults.All() {
		assert.Equal(t, "default", setting.Subject, setting.Identifier)
	}
}

// A level stores only what it has set, so the level below shows through for the rest, and a setting turned off is
// stored as off, not left out
func TestALevelStoresOnlyWhatItSets(t *testing.T) {
	site := NewLevel(map[string]any{"timezone": "Pacific/Auckland", "alarm_low_enabled": true, "range_low_mgdl": 80})
	identity := NewLevel(map[string]any{"alarm_low_enabled": false})
	stored, err := site.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland","range":{"low_mgdl":80},"alarms":{"low":{"enabled":true}}}`, stored)
	stored, err = identity.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"alarms":{"low":{"enabled":false}}}`, stored)

	s := FromLevels(LevelFromJSON(must(site.JSON()), "site", 1), LevelFromJSON(stored, "identity", 2))
	assert.True(t, s.Site.Alarms.Low.Enabled, "the site's own")
	assert.False(t, s.Alarms.Low.Enabled, "the identity turned it off for itself")
	assert.Equal(t, int64(80), s.Range.LowMgdl, "the site's, as the identity has not set one")
	assert.Equal(t, int64(72), s.Default.Range.LowMgdl)
	assert.Equal(t, "mmol/L", string(s.Units), "the default, as neither has set one")
	low, err := s.Get("alarm_low_enabled")
	require.NoError(t, err)
	assert.Equal(t, "identity", low.Subject)
	timezone, err := s.Get("timezone")
	require.NoError(t, err)
	assert.Equal(t, "site", timezone.Subject)

	// A row with nothing set, as a new site's and a new identity's, has no settings of its own
	stored, err = LevelFromJSON(`{}`, "site", 0).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1}`, stored)
}

// A change to a level shows once the settings are rebuilt, and a setting cleared goes back to the level below
func TestRebuildFollowsALevel(t *testing.T) {
	s := FromLevels(NewLevel(map[string]any{"units": "mg/dL"}), nil)
	require.NoError(t, s.IdentityLevel.Set("units", "mmol/L"))
	assert.Equal(t, "mg/dL", string(s.Units), "not until rebuilt")
	assert.Equal(t, "mmol/L", string(s.Rebuild().Units))
	require.NoError(t, s.IdentityLevel.Set("units", nil))
	assert.Equal(t, "mg/dL", string(s.Rebuild().Units))
}

// A stored row is not trusted: a value out of its range, not one of its options or of the wrong type is left out and
// the rest kept, a key that is not a field is not read, and a row that is not JSON or is too long has no settings
func TestARowIsValidatedWhenRead(t *testing.T) {
	stored, err := LevelFromJSON(`{"version":1,"timezone":"Pacific/Auckland","units":"stone","range":{"low_mgdl":5000,"high_mgdl":"high","urgent_high_mgdl":300},"alarms":[1],"is_admin":true}`, "site", 0).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland","range":{"urgent_high_mgdl":300}}`, stored)

	for name, row := range map[string]string{"not JSON": `{"timezone":`, "not an object": `["timezone"]`, "empty": ``,
		"too long": `{"timezone":"Pacific/Auckland","pad":"` + strings.Repeat("x", MaxJSONBytes) + `"}`} {
		stored, err := LevelFromJSON(row, "site", 0).JSON()
		require.NoError(t, err, name)
		assert.JSONEq(t, `{"version":1}`, stored, name)
	}
}

// Every setting set, each as long as it may be, fits its column with room to spare, so the column's limit is never
// what refuses a setting
func TestALevelFitsItsColumn(t *testing.T) {
	level := LevelFromJSON(valuesV1, "site", 0)
	devices := make([]any, maxDeviceNames)
	for i := range devices {
		devices[i] = strings.Repeat(string(rune('a'+i)), maxDeviceName)
	}
	require.NoError(t, level.Set("alarm_device_battery_low_devices", devices))
	stored, err := level.JSON()
	require.NoError(t, err)
	assert.Less(t, len(stored), MaxJSONBytes/2, fmt.Sprint(len(stored)))
}

func must(stored string, err error) string {
	if err != nil {
		panic(err)
	}
	return stored
}

// BenchmarkSettingsNew is building a request's settings from a site's row and an identity's, which a request does
// once: as most are, a site with a few settings and an identity with one, and at most, each with every setting
// Run with go test -run XXX -bench SettingsNew ./utils/settings
func BenchmarkSettingsNew(b *testing.B) {
	site := must(NewLevel(map[string]any{"timezone": "Pacific/Auckland", "language": "en-NZ", "units": "mmol/L", "date_format": "DD/MM/YYYY",
		"time_format": "hh:mm a", "week_start_day": "monday", "alarm_low_enabled": true}).JSON())
	identity := must(NewLevel(map[string]any{"units": "mg/dL"}).JSON())
	for name, rows := range map[string][2]string{"usual": {site, identity}, "every setting": {valuesV1, valuesV1}} {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				if FromLevels(LevelFromJSON(rows[0], "site", 0), LevelFromJSON(rows[1], "identity", 0)).Timezone != "Pacific/Auckland" {
					b.Fatal("settings not read")
				}
			}
		})
	}
}

// What version 1 stored is fixed: its example row, and the fields it had, the first lines of values_schema.txt. A
// change to either fails here, so they cannot be edited to make the tests above pass. Fields added since version 1 are
// lines after these.
const (
	version1Fields = 46
	version1Row    = "24e599c57f23cfb83101ccad88cbcf8c881823e6439260fd69fbdb292cab3906"
	version1Schema = "8adf68fd3195f357237ad5b017d5674ac165f0f8cacdabaea1bac300d9441681"
)

func TestWhatVersion1StoredIsNotEdited(t *testing.T) {
	assert.Equal(t, version1Row, fmt.Sprintf("%x", sha256.Sum256([]byte(valuesV1))), "values_v1.json was changed: put it back")
	lines := strings.SplitAfter(valuesSchema, "\n")
	require.GreaterOrEqual(t, len(lines), version1Fields)
	assert.Equal(t, version1Schema, fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(lines[:version1Fields], "")))),
		"a line of values_schema.txt that version 1 had was changed: put it back")
}
