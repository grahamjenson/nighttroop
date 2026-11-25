package settings

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
)

// MaxJSONBytes is the most a level's settings take as stored, as their column allows. Every setting set, with the
// longest values each allows (10 device names of 64 characters), is under half of it.
const MaxJSONBytes = 4096

// Level is the settings one level has set of its own: the site's, or an identity's. A setting it has not set is not
// in it, so the level below, the site's or the default, shows through.
type Level struct {
	// Subject is whose the level is, "site" or "identity", and At when it was made
	Subject string
	At      int64
	// own is each setting the level has set, by its identifier, as the setting's own type
	own map[string]any
}

// NewLevel is a level with the values, each by its identifier and parsed and validated as one sent to the API is: a
// nil value, or one that is not valid, is left out
func NewLevel(values map[string]any) *Level {
	l := &Level{own: map[string]any{}}
	for name, value := range values {
		// A pointer is the value it points to, and nil is not set
		if at := reflect.ValueOf(value); at.Kind() == reflect.Pointer {
			if at.IsNil() {
				continue
			}
			value = at.Elem().Interface()
		}
		_ = l.Set(name, value)
	}
	return l
}

// LevelFromJSON is the level a settings_json column holds. Each value is parsed and validated as one sent to the API
// is, so a value out of its range or of the wrong type, or a key that is not a setting, is left out and the rest
// read: a row is never trusted because it was stored. A row that is too long, or is not JSON, has no settings.
func LevelFromJSON(raw string, subject string, at int64) *Level {
	l := &Level{Subject: subject, At: at, own: map[string]any{}}
	var stored map[string]any
	// A row made in code and not read from its table has none
	if raw == "" {
		return l
	}
	if len(raw) > MaxJSONBytes {
		slog.Error("stored settings are too long, so none are used", "bytes", len(raw))
		return l
	}
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		slog.Error("stored settings cannot be read, so none are used", "error", err)
		return l
	}
	for _, f := range allSettings {
		if value, ok := at2(stored, f.path); ok && value != nil {
			if err := l.Set(f.identifier, value); err != nil {
				slog.Warn("a stored setting is not valid, so it is not used", "setting", f.identifier)
			}
		}
	}
	return l
}

// at2 is the value at a path of nested JSON objects
func at2(object map[string]any, path []string) (any, bool) {
	for _, key := range path[:len(path)-1] {
		next, ok := object[key].(map[string]any)
		if !ok {
			return nil, false
		}
		object = next
	}
	value, ok := object[path[len(path)-1]]
	return value, ok
}

// JSON is the level as its column holds it: a Values with only the settings the level has set, at Version
func (l *Level) JSON() (string, error) {
	stored := map[string]any{"version": Version}
	for _, f := range allSettings {
		value, ok := l.own[f.identifier]
		if !ok {
			continue
		}
		object := stored
		for _, key := range f.path[:len(f.path)-1] {
			next, ok := object[key].(map[string]any)
			if !ok {
				next = map[string]any{}
				object[key] = next
			}
			object = next
		}
		object[f.path[len(f.path)-1]] = value
	}
	raw, err := json.Marshal(stored)
	if err == nil && len(raw) > MaxJSONBytes {
		err = fmt.Errorf("settings are %d bytes, over %d", len(raw), MaxJSONBytes)
	}
	return string(raw), err
}

// Set sets a setting of the level to a value sent or stored, parsed and validated, or with nil clears it so the level
// below shows through. A value that is not valid, or an identifier that is not a setting, changes nothing.
func (l *Level) Set(identifier string, value any) error {
	f, ok := byName[identifier]
	if !ok {
		return nil
	}
	if value == nil {
		delete(l.own, identifier)
		return nil
	}
	parsed, err := f.parse(value)
	if err != nil {
		return err
	}
	l.own[identifier] = parsed
	return nil
}

// Get is the level's own value of a setting, and whether it has set one
func (l *Level) Get(identifier string) (any, bool) {
	value, ok := l.own[identifier]
	return value, ok
}

// Setting is the level's own value of a setting as the APIs give it, or nil when it has not set one
func (l *Level) Setting(identifier string) *SettingAnyType {
	value, ok := l.own[identifier]
	if !ok {
		return nil
	}
	return &SettingAnyType{Identifier: identifier, TypeString: byName[identifier].typeString, Value: value, CreatedAt: l.At, Subject: l.Subject}
}

// applyTo sets each setting the level has set in v
func (l *Level) applyTo(v *Values) {
	if len(l.own) == 0 {
		return
	}
	for _, f := range v.fields() {
		if value, ok := l.own[f.identifier]; ok {
			f.set(value)
		}
	}
}

// StoredOf are the values as a column takes them, as NewLevel reads them
func StoredOf(values map[string]any) string {
	stored, _ := NewLevel(values).JSON()
	return stored
}
