package settings

import "reflect"

func first[T any](values ...*T) *T {
	for _, value := range values {
		if value == nil {
			continue
		}

		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
			if v.IsNil() {
				continue
			}
		}

		return value
	}

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
