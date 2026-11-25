package settings

import (
	"reflect"
)

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

// Of are the settings with the values, each by its identifier and parsed and validated as one sent to the API is
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
