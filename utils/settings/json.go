package settings

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	return of(values, createdAt, subject)
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
