package nsutils

import (
	"fmt"
	"strings"
)

type Collection string

const (
	Entries      Collection = "entries"
	Treatments   Collection = "treatments"
	Devicestatus Collection = "devicestatus"
	Profile      Collection = "profile"
	Properties   Collection = "properties"
	Food         Collection = "food"
	Settings     Collection = "settings"
)

func ParseCollection(s string) (Collection, error) {

	// Validate and normalize
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "device_status":
		s = "devicestatus"
	case "profiles":
		s = "profile"
	case "foods":
		s = "food"
	}

	c := Collection(s)
	switch c {
	case Entries,
		Treatments,
		Devicestatus,
		Profile,
		Properties,
		Food,
		Settings:
		return c, nil
	default:
		return "", fmt.Errorf("unknown collection %q", s)
	}
}

func (c Collection) DefaultDateField() string {
	switch c {
	case Entries:
		return "date"
	default:
		return "created_at"
	}
}
