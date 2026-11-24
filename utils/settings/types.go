package settings

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/i18n"
)

type SettingAnyType struct {
	Identifier string
	TypeString string
	Value      any
	CreatedAt  int64
	Subject    string
}

type Setting[T any] struct {
	definition *SettingDefinition[T]
	Value      T
	CreatedAt  int64
	Subject    string
}

func (s *Setting[T]) ToSettingAnyType() *SettingAnyType {
	if s == nil {
		return nil
	}
	return &SettingAnyType{
		Identifier: s.definition.Identifier,
		TypeString: s.definition.TypeString,
		Value:      s.Value,
		CreatedAt:  s.CreatedAt,
		Subject:    s.Subject,
	}
}

func NewSetting[T any](definition *SettingDefinition[T], value *T, createdAt *int64, subject *string) *Setting[T] {
	// Setting value to nil is just deleting the setting
	if value == nil {
		return nil
	}

	if createdAt == nil {
		now := time.Now().UnixMilli()
		createdAt = &now
	}

	if subject == nil {
		subject = ptr("default")
	}

	setting := &Setting[T]{
		definition: definition,
		Value:      *value,
		CreatedAt:  *createdAt,
		Subject:    *subject,
	}

	return setting
}

func NewDefaultSetting[T any](definition *SettingDefinition[T]) *Setting[T] {

	setting := &Setting[T]{
		definition: definition,
		Value:      definition.Default,
		CreatedAt:  time.Now().UnixMilli(),
		Subject:    "default",
	}

	return setting
}

func NewSettingAndParseValidate[T any](definition *SettingDefinition[T], value any, createdAt *int64, subject *string) (*Setting[T], error) {
	if value == nil {
		return nil, nil
	}
	parsedValue, err := definition.ParseValidate(value)
	if err != nil {
		return nil, err
	}
	return NewSetting(definition, parsedValue, createdAt, subject), nil
}

// //
// Parsers
// ////
func ParseUnits(value any) (consts.Units, error) {
	switch v := value.(type) {
	case *consts.Units:
		return *v, nil
	case consts.Units:
		return v, nil
	case string:
		u, err := consts.ParseUnits(v)
		if err != nil {
			return consts.UnitsMMOLL, fmt.Errorf("cannot parse units from string: %v", err)
		}
		return u, nil
	default:
		return consts.UnitsMMOLL, fmt.Errorf("unsupported type for ParseUnits: %T", value)
	}
}

func ParseInt(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		var i int64
		_, err := fmt.Sscanf(v, "%d", &i)
		if err != nil {
			return 0, fmt.Errorf("cannot parse int from string: %v", err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported type for ParseInt: %T", value)
	}
}

func ParseFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return 0, fmt.Errorf("cannot parse float from string: %v", err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported type for ParseFloat: %T", value)
	}
}

func ParseBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "true" {
			return true, nil
		} else if v == "false" {
			return false, nil
		} else {
			return false, fmt.Errorf("cannot parse bool from string: %v", v)
		}
	default:
		return false, fmt.Errorf("unsupported type for ParseBool: %T", value)
	}
}

func ParseString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported type for ParseString: %T", value)
	}
}

func ParseLowerString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return strings.ToLower(v), nil
	default:
		return "", fmt.Errorf("unsupported type for ParseLowerString: %T", value)
	}
}

// ParseDeviceNames is a list of device names, as a JSON array is decoded
func ParseDeviceNames(value any) (types.DeviceNames, error) {
	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("unsupported type for ParseDeviceNames: %T", value)
	}
	names := make(types.DeviceNames, len(list))
	for i, item := range list {
		name, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("device names must be strings, not %T", item)
		}
		names[i] = name
	}
	return names, nil
}

/////
// Validators
/////

// maxDeviceNames and maxDeviceName keep the list of device names within its column
const maxDeviceNames, maxDeviceName = 10, 64

// ValidDeviceNames is at most 10 different device names, each of 1 to 64 characters
func ValidDeviceNames(names types.DeviceNames) error {
	if len(names) > maxDeviceNames {
		return i18n.NewError("errors.settings.too_many_devices", i18n.Data{"Max": maxDeviceNames})
	}
	for i, name := range names {
		if name == "" || len(name) > maxDeviceName {
			return i18n.NewError("errors.settings.device_length", i18n.Data{"Max": maxDeviceName})
		}
		if slices.Contains(names[:i], name) {
			return i18n.NewError("errors.settings.device_twice", i18n.Data{"Name": name})
		}
	}
	return nil
}

func Range[T cmp.Ordered](min, max T) func(T) error {
	return func(value T) error {
		if value < min || value > max {
			return i18n.NewError("errors.settings.range", i18n.Data{"Min": min, "Max": max})
		}
		return nil
	}
}

func quoted(values []string) string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = fmt.Sprintf("'%v'", value)
	}
	return strings.Join(result, ", ")
}
