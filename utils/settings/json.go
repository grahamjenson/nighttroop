package settings

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
)

// MaxJSONBytes is the most a level's settings take as stored. Every setting set, with the longest values each
// allows (10 device names of 64 characters), is under 2 KiB, so this is only reached by a row not written here.
const MaxJSONBytes = 4096

// JSON is the settings m has set, each by its identifier, as one stored row holds them. A setting that is not set is
// left out, so the level below it, the site's or the default, shows through.
func (m *SettingMap) JSON() ([]byte, error) {
	values := map[string]any{}
	for _, s := range m.slots() {
		if setting := s.anyType(); setting != nil {
			values[s.identifier()] = setting.Value
		}
	}
	raw, err := json.Marshal(values)
	if err == nil && len(raw) > MaxJSONBytes {
		err = fmt.Errorf("settings are %d bytes, over %d", len(raw), MaxJSONBytes)
	}
	return raw, err
}

// Stored is JSON as a column takes it, for a save's parameters
func (m *SettingMap) Stored() (string, error) {
	raw, err := m.JSON()
	return string(raw), err
}

// StoredOf are the values as a column takes them, each by its identifier: a nil value, or one not valid, is left out
func StoredOf(values map[string]any) string {
	for name, value := range values {
		if v := reflect.ValueOf(value); v.Kind() == reflect.Pointer {
			delete(values, name)
			if !v.IsNil() {
				values[name] = v.Elem().Interface()
			}
		}
	}
	stored, _ := Of(values).Stored()
	return stored
}

// FromJSON is the settings a stored row holds. Each value is parsed and validated as one sent to the API is, so a key
// that is not a setting, a value of the wrong type or out of its range, or a row too long or not JSON, is left out:
// a row is never trusted because it was stored.
func FromJSON(raw []byte, createdAt *int64, subject *string) *SettingMap {
	var values map[string]any
	if len(raw) > MaxJSONBytes || json.Unmarshal(raw, &values) != nil {
		return &SettingMap{}
	}
	m := of(values, createdAt, subject)
	for name := range values {
		if m.GetSettingAnyTypeByName(name) == nil {
			warnLeftOut(name)
		}
	}
	return m
}

// leftOut are the stored settings already logged as left out, so each is logged once and not on every request
var leftOut sync.Map

// warnLeftOut logs that a stored value is not used: its setting is gone, or the value is no longer valid, after a
// change to the settings in code. The value stays in its row, as a save changes only the settings it sets, so a
// release that left it out by mistake can be put right with nothing lost. The value is not logged.
func warnLeftOut(name string) {
	if _, logged := leftOut.LoadOrStore(name, true); !logged {
		slog.Warn("A stored setting is not valid and is left out, so the level below it is used", "setting", name)
	}
}

// Changes are what differs between a level as its row has it and as m now has it: a JSON merge patch for the row,
// with each changed setting by its identifier and null for one no longer set, and the changed settings' names in the
// order the settings are listed. A save writes only this, so two saves at once do not lose each other's settings,
// and a key of the row that is not a setting now is left as it is.
func (m *SettingMap) Changes(stored string) (patch string, names []string, err error) {
	before, now := FromJSON([]byte(stored), nil, nil), map[string]any{}
	for _, s := range m.slots() {
		was, is := s.of(before), s.anyType()
		switch {
		case is == nil && was != nil:
			now[s.identifier()] = nil
		case is != nil && (was == nil || !reflect.DeepEqual(was.Value, is.Value)):
			now[s.identifier()] = is.Value
		default:
			continue
		}
		names = append(names, s.identifier())
	}
	raw, err := json.Marshal(now)
	return string(raw), names, err
}

// Of are the settings with the values, each by its identifier and parsed and validated as FromJSON does
func Of(values map[string]any) *SettingMap {
	return of(values, nil, nil)
}

func of(values map[string]any, createdAt *int64, subject *string) *SettingMap {
	m := &SettingMap{}
	for _, s := range m.slots() {
		if value, ok := values[s.identifier()]; ok {
			_ = s.parseValidate(value, createdAt, subject)
		}
	}
	return m
}
