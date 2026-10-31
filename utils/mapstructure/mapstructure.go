// Package mapstructure decodes query parameters into a struct by its fields' mapstructure tags: the part of
// github.com/mitchellh/mapstructure, now archived, that Nighttroop used.
package mapstructure

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

var unmarshaler = reflect.TypeFor[json.Unmarshaler]()

// Decode sets the fields of result, a pointer to a struct, from values, matching each value's key to a field's
// mapstructure tag name with match, or case-insensitively when match is nil. A field is a *string, or a pointer to a
// type with its own UnmarshalJSON, given the value as a JSON string. A field tagged ",remain", a map[string]any, gets
// the values no field matched; without one they are left out.
func Decode(values map[string]string, result any, match func(key, name string) bool) error {
	if match == nil {
		match = strings.EqualFold
	}
	target := reflect.ValueOf(result).Elem()
	keys := slices.Sorted(func(yield func(string) bool) {
		for key := range values {
			if !yield(key) {
				return
			}
		}
	})
	used := map[string]bool{}
	var remain reflect.Value
	for i := range target.NumField() {
		field, value := target.Type().Field(i), target.Field(i)
		name, options, _ := strings.Cut(field.Tag.Get("mapstructure"), ",")
		if options == "remain" {
			remain = value
			continue
		}
		key, found := name, false
		if _, found = values[name]; !found {
			for _, candidate := range keys {
				if found = match(candidate, name); found {
					key = candidate
					break
				}
			}
		}
		if !found {
			continue
		}
		used[key] = true
		if err := set(value, values[key]); err != nil {
			return fmt.Errorf("'%s' %w", name, err)
		}
	}
	if remain.IsValid() {
		for _, key := range keys {
			if !used[key] {
				if remain.IsNil() {
					remain.Set(reflect.MakeMap(remain.Type()))
				}
				remain.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(values[key]))
			}
		}
	}
	return nil
}

// set sets field, a pointer, to text: a *string to it, and any other to the type's UnmarshalJSON of it
func set(field reflect.Value, text string) error {
	if field.Kind() != reflect.Pointer {
		return fmt.Errorf("must be a pointer, not %s", field.Type())
	}
	value := reflect.New(field.Type().Elem())
	switch {
	case field.Type().Elem().Kind() == reflect.String && !value.Type().Implements(unmarshaler):
		value.Elem().SetString(text)
	case value.Type().Implements(unmarshaler):
		data, _ := json.Marshal(text)
		if err := value.Interface().(json.Unmarshaler).UnmarshalJSON(data); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot be decoded into %s", field.Type())
	}
	field.Set(value)
	return nil
}
