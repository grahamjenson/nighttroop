package settings

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

// settings_schema.txt is every field of Settings: its place in the JSON, its type and its setting. It only gains
// lines.
//
//go:embed settings_schema.txt
var settingsSchema string

// settings_v1.json is a row as version 1 stored it, with every setting of that version set
//
//go:embed settings_v1.json
var settingsV1 string

// schema is each field of Settings as a line: its JSON path, its Go type and the setting kept in it, in the order
// they are declared
func schema() []string {
	settingAt := map[string]string{}
	for _, f := range allSettings {
		settingAt[f.path] = f.identifier
	}
	return schemaOf(reflect.TypeFor[Settings](), "", settingAt)
}

func schemaOf(t reflect.Type, path string, settingAt map[string]string) []string {
	var lines []string
	for i := range t.NumField() {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		at := strings.TrimPrefix(path+"."+name, ".")
		if t.Field(i).Type.Kind() == reflect.Struct {
			lines = append(lines, schemaOf(t.Field(i).Type, at, settingAt)...)
			continue
		}
		lines = append(lines, strings.TrimSpace(at+" "+t.Field(i).Type.String()+" "+settingAt[at]))
	}
	return lines
}

// Rows of every release are read into Settings, so a field is never renamed, removed, moved, given another type or
// given to another setting: its row would be read as not set, and the person given the default with no word. A new
// field is fine, and is added to settings_schema.txt in the same change. Never change or delete a line of that file to
// make this pass.
func TestSettingsOnlyGrow(t *testing.T) {
	now := schema()
	released := strings.Split(strings.TrimSpace(settingsSchema), "\n")
	for _, line := range released {
		assert.Contains(t, now, line, "a released field of Settings was renamed, removed, moved or retyped: put it back as it was")
	}
	for _, line := range now {
		assert.Contains(t, released, line, "a new field of Settings: add this line to the end of settings_schema.txt")
	}
	assert.Equal(t, len(released), len(slices.Compact(slices.Sorted(slices.Values(released)))), "a line is in settings_schema.txt twice")
}

// A row a release stored is read whole by every later one: each setting it has is read with the value it had, and
// written back the same. To change what a stored value means, add a version and read version 1 as before; never
// change settings_v1.json.
func TestARowOfVersion1IsStillRead(t *testing.T) {
	s := SettingsFromJSON(settingsV1)
	again, err := s.JSON()
	require.NoError(t, err)
	assert.JSONEq(t, settingsV1, again)
	for _, f := range allSettings {
		_, set := f.own(&s)
		assert.True(t, set, f.identifier)
	}

	// And is the values it says, in effect over the defaults
	v := (&ContextSettings{DefaultSettings: DefaultValues(), Site: s}).Values()
	assert.Equal(t, consts.UnitsMGDL, v.Units)
	assert.Equal(t, "Pacific/Auckland", v.Timezone)
	assert.Equal(t, int64(80), v.Range.LowMgdl)
	assert.Equal(t, int64(10), v.Alarms.Low.Minutes)
	assert.Equal(t, int64(20), v.Alarms.FastRise.Minutes)
	assert.Equal(t, int64(10), v.Alarms.FastDrop.Minutes)
	assert.Equal(t, 0.8, v.Defaults.BasalRate)
	assert.Equal(t, types.DeviceNames{"openaps://pixel", "SugarPixel"}, v.Alarms.DeviceBatteryLow.Devices)
}

// Every setting has one field of Settings and the same field of SettingsValues, and every field is a setting, so none
// is lost when saved, and with nothing set every one is at its default
func TestEverySettingHasItsField(t *testing.T) {
	assert.Len(t, schema(), len(names)+1, "a field of Settings that is no setting's, but for version")
	assert.Len(t, slices.Compact(slices.Sorted(slices.Values(names))), len(names), "a setting is in two fields")
	for _, line := range schema()[1:] {
		assert.Len(t, strings.Fields(line), 3, "a field with no setting: "+line)
	}
	values := map[string][]int{}
	placesOf(reflect.TypeFor[SettingsValues](), "", nil, values)
	assert.Len(t, values, len(names), "SettingsValues has a field Settings has not")

	defaults := &ContextSettings{DefaultSettings: DefaultValues()}
	assert.Equal(t, DefaultValues(), defaults.Values())
	assert.Equal(t, "UTC", defaults.GetTimezone())
	assert.Equal(t, consts.UnitsMMOLL, defaults.GetUnits())
	assert.Equal(t, int64(72), defaults.GetRange().LowMgdl)
	assert.False(t, defaults.Values().Alarms.Low.Enabled)
	for _, setting := range defaults.All() {
		assert.Equal(t, "default", setting.Subject, setting.Identifier)
	}
}

// A Settings stores only what is set, so the level below shows through for the rest, and a setting turned off is
// stored as off, not left out
func TestSettingsStoreOnlyWhatIsSet(t *testing.T) {
	site := Settings{Timezone: new("Pacific/Auckland"), Range: GlucoseRange{LowMgdl: new(int64(80))}, Alarms: Alarms{Low: TimedAlarm{Enabled: new(true)}}}
	identity := Settings{Alarms: Alarms{Low: TimedAlarm{Enabled: new(false)}}}
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland","range":{"low_mgdl":80},"alarms":{"low":{"enabled":true}}}`, StoredOf(site))
	assert.JSONEq(t, `{"version":1,"alarms":{"low":{"enabled":false}}}`, StoredOf(identity))

	c := &ContextSettings{DefaultSettings: DefaultValues(), Site: SettingsFromJSON(StoredOf(site)), Identity: SettingsFromJSON(StoredOf(identity)), SiteSince: 1, IdentitySince: 2}
	assert.False(t, c.Values().Alarms.Low.Enabled, "the identity turned it off for itself")
	assert.True(t, *c.Site.Alarms.Low.Enabled, "the site's own")
	assert.Equal(t, int64(80), c.GetRange().LowMgdl, "the site's, as the identity has not set one")
	assert.Equal(t, int64(72), c.DefaultSettings.Range.LowMgdl)
	assert.Equal(t, consts.UnitsMMOLL, c.GetUnits(), "the default, as neither has set one")
	assert.Equal(t, "Pacific/Auckland", c.GetSiteTimezone())
	low, err := c.Get("alarm_low_enabled")
	require.NoError(t, err)
	assert.Equal(t, "identity", low.Subject)
	assert.Equal(t, int64(2), low.CreatedAt)
	timezone, err := c.Get("timezone")
	require.NoError(t, err)
	assert.Equal(t, "site", timezone.Subject)

	// A row with nothing set, as a new site's and a new identity's, has no settings of its own
	assert.JSONEq(t, `{"version":1}`, StoredOf(SettingsFromJSON(`{}`)))
}

// A setting is set by its name with a value as it is sent, and cleared with nil so the level below shows through
func TestSetAndClearByName(t *testing.T) {
	c := &ContextSettings{DefaultSettings: DefaultValues(), Site: Settings{Units: new(consts.UnitsMGDL)}}
	require.NoError(t, c.Identity.Set("units", "mmol/L"))
	require.NoError(t, c.Identity.Set("alarm_low_minutes", "15"))
	require.NoError(t, c.Identity.Set("alarm_device_battery_low_devices", []any{"phone"}))
	assert.Equal(t, consts.UnitsMMOLL, c.GetUnits())
	assert.Equal(t, int64(15), *c.Identity.Alarms.Low.Minutes)
	assert.Equal(t, types.DeviceNames{"phone"}, c.Values().Alarms.DeviceBatteryLow.Devices)
	require.Error(t, c.Identity.Set("alarm_low_minutes", "-1"))
	assert.Equal(t, int64(15), *c.Identity.Alarms.Low.Minutes, "a value that is not valid changes nothing")
	require.NoError(t, c.Identity.Set("units", nil))
	assert.Equal(t, consts.UnitsMGDL, c.GetUnits())
	c.SetIdentityUnits(consts.UnitsMMOLL)
	assert.Equal(t, consts.UnitsMMOLL, c.GetUnits())
}

// A stored row is not trusted: a value out of its range, not one of its options or of the wrong type is left out and
// the rest kept, a key that is not a field is not read, and a row that is not JSON or is too long has no settings
func TestARowIsValidatedWhenRead(t *testing.T) {
	row := `{"version":1,"timezone":"Pacific/Auckland","units":"stone","range":{"low_mgdl":5000,"high_mgdl":"high","urgent_high_mgdl":300},"alarms":[1],"is_admin":true}`
	assert.JSONEq(t, `{"version":1,"timezone":"Pacific/Auckland","range":{"urgent_high_mgdl":300}}`, StoredOf(SettingsFromJSON(row)))

	for name, row := range map[string]string{"not JSON": `{"timezone":`, "not an object": `["timezone"]`, "empty": ``,
		"too long": `{"timezone":"Pacific/Auckland","pad":"` + strings.Repeat("x", MaxJSONBytes) + `"}`} {
		assert.JSONEq(t, `{"version":1}`, StoredOf(SettingsFromJSON(row)), name)
	}
}

// Every setting set, each as long as it may be, fits its column with room to spare, so the column's limit is never
// what refuses a setting
func TestSettingsFitTheirColumn(t *testing.T) {
	s := SettingsFromJSON(settingsV1)
	devices := make(types.DeviceNames, maxDeviceNames)
	for i := range devices {
		devices[i] = strings.Repeat(string(rune('a'+i)), maxDeviceName)
	}
	s.Alarms.DeviceBatteryLow.Devices = &devices
	stored, err := s.JSON()
	require.NoError(t, err)
	assert.Less(t, len(stored), MaxJSONBytes/2, fmt.Sprint(len(stored)))
}

// BenchmarkContextSettings is what a request pays for its settings: reading a site's row and an identity's once, as
// most are and with every setting set, and then a getter each time a setting is read
// Run with go test -run XXX -bench ContextSettings ./utils/settings
func BenchmarkContextSettings(b *testing.B) {
	usual := [2]string{StoredOf(Settings{Timezone: new("Pacific/Auckland"), Language: new("en-NZ"), Units: new(consts.UnitsMMOLL), DateFormat: new("DD/MM/YYYY"),
		TimeFormat: new("hh:mm a"), WeekStartDay: new("monday"), Alarms: Alarms{Low: TimedAlarm{Enabled: new(true)}}}), StoredOf(Settings{Units: new(consts.UnitsMGDL)})}
	for name, rows := range map[string][2]string{"read, usual": usual, "read, every setting": {settingsV1, settingsV1}} {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				c := ContextSettings{DefaultSettings: DefaultValues(), Site: SettingsFromJSON(rows[0]), Identity: SettingsFromJSON(rows[1])}
				if c.GetTimezone() != "Pacific/Auckland" {
					b.Fatal("settings not read")
				}
			}
		})
	}
	c := &ContextSettings{DefaultSettings: DefaultValues(), Site: SettingsFromJSON(usual[0]), Identity: SettingsFromJSON(usual[1])}
	b.Run("get units", func(b *testing.B) {
		for b.Loop() {
			if c.GetUnits() != consts.UnitsMGDL {
				b.Fatal("not the identity's units")
			}
		}
	})
	b.Run("every value", func(b *testing.B) {
		for b.Loop() {
			if c.Values().Timezone != "Pacific/Auckland" {
				b.Fatal("settings not read")
			}
		}
	})
}

// What version 1 stored is fixed: its example row, and the fields it had, the first lines of settings_schema.txt. A
// change to either fails here, so they cannot be edited to make the tests above pass. Fields added since version 1 are
// lines after these.
const (
	version1Fields = 46
	version1Row    = "24e599c57f23cfb83101ccad88cbcf8c881823e6439260fd69fbdb292cab3906"
	version1Schema = "9a36625530e0b87f385577b5c010bf6e392ad4e6d5e4ad0f9efa8f291f7e68d1"
)

func TestWhatVersion1StoredIsNotEdited(t *testing.T) {
	assert.Equal(t, version1Row, fmt.Sprintf("%x", sha256.Sum256([]byte(settingsV1))), "settings_v1.json was changed: put it back")
	lines := strings.SplitAfter(settingsSchema, "\n")
	require.GreaterOrEqual(t, len(lines), version1Fields)
	assert.Equal(t, version1Schema, fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(lines[:version1Fields], "")))),
		"a line of settings_schema.txt that version 1 had was changed: put it back")
}
