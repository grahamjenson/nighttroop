package types

import (
	"encoding/json"
	"reflect"
)

// JSONUnmarshalerHook is a mapstructure decode hook that decodes into any type with its own UnmarshalJSON
func JSONUnmarshalerHook(
	from reflect.Type,
	to reflect.Type,
	data any,
) (any, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

	// We need a pointer to the destination type because
	// UnmarshalJSON is usually defined on *T.
	ptrType := reflect.PointerTo(to)
	if !ptrType.Implements(unmarshalerType) {
		return data, nil
	}

	// Marshal the incoming value back to JSON.
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create a new instance of T.
	v := reflect.New(to)

	// Call UnmarshalJSON.
	if err := v.Interface().(json.Unmarshaler).UnmarshalJSON(b); err != nil {
		return nil, err
	}

	return v.Elem().Interface(), nil
}
