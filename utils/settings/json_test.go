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

// stored_schema.txt is every field of Stored: its place in the JSON, its type and its setting. It only gains lines.
//
//go:embed stored_schema.txt
var storedSchema string

// stored_v1.json is a row as version 1 stored it, with every setting of that version set
//
//go:embed stored_v1.json
var storedV1 string

// schema is each field of Stored as a line: its JSON path, its Go type and the setting kept in it, in the order
// they are declared
func schema() []string {
	s := &Stored{}
	settingAt := map[any]string{}
	for _, field := range s.fields() {
		settingAt[field.at] = field.identifier
	}
	return schemaOf(reflect.ValueOf(s).Elem(), "", settingAt)
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

// Rows of every release are read into Stored, so a field is never renamed, removed, moved, given another type or
// given to another setting:
// its row would be read as not set, and the person given the default with no word. A new field is fine, and is
// added to stored_schema.txt in the same change. Never change or delete a line of that file to make this pass.
func TestStoredOnlyGrows(t *testing.T) {
	now := schema()
	released := strings.Split(strings.TrimSpace(storedSchema), "\n")
	for _, line := range released {
		assert.Contains(t, now, line, "a released field of Stored was renamed, removed, moved or retyped: put it back as it was")
	}
	for _, line := range now {
		assert.Contains(t, released, line, "a new field of Stored: add this line to the end of stored_schema.txt")
	}
	assert.Equal(t, len(released), len(slices.Compact(slices.Sorted(slices.Values(released)))), "a line is in stored_schema.txt twice")
}

// A row a release stored is read whole by every later one: each setting it has is read with the value it had, and
// written back the same. To change what a stored value means, add a version and read version 1 as before; never
// change stored_v1.json.
func TestAStoredRowOfVersion1IsStillRead(t *testing.T) {
	read := FromJSON([]byte(storedV1), nil, nil)
	again, err := read.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, storedV1, string(again))
	assert.Equal(t, "Pacific/Auckland", *read.GetTimezone())
	assert.Equal(t, int64(80), *read.GetRangeLowMgdl())
	assert.Equal(t, types.DeviceNames{"openaps://pixel", "SugarPixel"}, *read.GetAlarmDeviceBatteryLowDevicesJson())
}

// Every setting has one place in Stored, and every place is a setting, so none is lost when saved
func TestEverySettingIsStored(t *testing.T) {
	var stored []string
	for _, field := range (&Stored{}).fields() {
		stored = append(stored, field.identifier)
	}
	assert.ElementsMatch(t, allSettingNames, stored)
	assert.Len(t, schema(), len(stored)+1, "a field of Stored that is no setting's, but for version")

	// With every setting set, the defaults, nothing is lost and the version is written
	raw, err := DefaultSettingsMap.JSON()
	require.NoError(t, err)
	var written Stored
	require.NoError(t, json.Unmarshal(raw, &written))
	assert.Equal(t, StoredVersion, written.Version)
	for _, field := range written.fields() {
		assert.NotNil(t, field.get(), field.identifier)
	}
	again, err := FromJSON(raw, nil, nil).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, string(raw), string(again))
}

// A level with only some settings set stores only those, so the level below shows through for the rest
func TestStoredLeavesOutWhatIsNotSet(t *testing.T) {
	m := Of(map[string]any{"timezone": "Pacific/Auckland", "alarm_low_enabled": true})
	raw, err := m.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland","alarms":{"low":{"enabled":true}}}`, string(raw))
	read := FromJSON(raw, nil, nil)
	assert.Equal(t, "Pacific/Auckland", *read.GetTimezone())
	assert.Nil(t, read.GetUnits())

	// A row with nothing set, as a new site's and a new identity's, has no settings
	raw, err = FromJSON([]byte(`{}`), nil, nil).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1}`, string(raw))
}

// A stored row is not trusted: a value out of its range or not one of its options is left out and the rest kept, a
// key that is not a field is not read, and a row that is not a Stored or is too long has no settings
func TestStoredIsValidatedWhenRead(t *testing.T) {
	raw, err := FromJSON([]byte(`{"version":1,"timezone":"Pacific/Auckland","units":"stone","range":{"low_mgdl":5000},"is_admin":true}`), nil, nil).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland"}`, string(raw))

	raw, err = FromJSON([]byte(`{"timezone":"Pacific/Auckland","range":{"low_mgdl":"low","high_mgdl":200},"alarms":[1]}`), nil, nil).JSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland","range":{"high_mgdl":200}}`, string(raw), "a field of another type")

	for name, row := range map[string]string{"not JSON": `{"timezone":`, "not an object": `["timezone"]`, "empty": ``,
		"too long": `{"timezone":"Pacific/Auckland","pad":"` + strings.Repeat("x", MaxJSONBytes) + `"}`} {
		raw, err := FromJSON([]byte(row), nil, nil).JSON()
		require.NoError(t, err, name)
		assert.JSONEq(t, `{"version":1}`, string(raw), name)
	}
}

// Every setting set, each as long as it may be, fits its column with room to spare, so the column's limit is never
// what refuses a setting
func TestStoredFitsItsColumn(t *testing.T) {
	m := FromJSON(must(DefaultSettingsMap.JSON()), nil, nil)
	devices := make(types.DeviceNames, maxDeviceNames)
	for i := range devices {
		devices[i] = strings.Repeat(string(rune('a'+i)), maxDeviceName)
	}
	m.SetAlarmDeviceBatteryLowDevicesJson(&devices)
	raw, err := m.JSON()
	require.NoError(t, err)
	assert.Less(t, len(raw), MaxJSONBytes/2, fmt.Sprint(len(raw)))
}

func must(raw []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return raw
}

// BenchmarkSettingsFromJSON is reading one level with every setting set
// Run with go test -run XXX -bench SettingsFromJSON ./utils/settings
func BenchmarkSettingsFromJSON(b *testing.B) {
	stored := must(DefaultSettingsMap.JSON())
	for b.Loop() {
		if FromJSON(stored, nil, nil).GetTimezone() == nil {
			b.Fatal("settings not read")
		}
	}
}

// What version 1 stored is fixed: its example row, and the fields it had, the first lines of stored_schema.txt. A
// change to either fails here, so they cannot be edited to make the tests above pass. Fields added since version 1 are
// lines after these.
const (
	version1Fields = 46
	version1Row    = "24e599c57f23cfb83101ccad88cbcf8c881823e6439260fd69fbdb292cab3906"
	version1Schema = "598d16ffd4a7deb172eeb844040419c948ad6057f4fc41299a53724692883663"
)

func TestWhatVersion1StoredIsNotEdited(t *testing.T) {
	assert.Equal(t, version1Row, fmt.Sprintf("%x", sha256.Sum256([]byte(storedV1))), "stored_v1.json was changed: put it back")
	lines := strings.SplitAfter(storedSchema, "\n")
	require.GreaterOrEqual(t, len(lines), version1Fields)
	assert.Equal(t, version1Schema, fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(lines[:version1Fields], "")))),
		"a line of stored_schema.txt that version 1 had was changed: put it back")
}
