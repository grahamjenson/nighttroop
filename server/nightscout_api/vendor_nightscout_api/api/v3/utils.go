package v3

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func validateCreateDoc(collection nsutils.Collection, doc map[string]any) error {

	if _, ok := nsutils.Int64Field(doc, "date"); !ok && collection != nsutils.Settings {
		return fmt.Errorf("Bad or missing date field")
	}
	if date, ok := nsutils.Int64Field(doc, "date"); ok && date <= 100000000000 {
		return fmt.Errorf("Bad or missing date field")
	}
	if utcOffset, exists := doc["utcOffset"]; exists {
		if value, err := toInt64Compat(utcOffset); err != nil || value < -1440 || value > 1440 {
			return fmt.Errorf("Bad or missing utcOffset field")
		}
	}
	if collection != nsutils.Settings {
		if app, ok := nsutils.StringField(doc, "app"); !ok || strings.TrimSpace(app) == "" {
			return fmt.Errorf("Bad or missing app field")
		}
	}
	return nil
}

func validateReplaceDoc(existing nsutils.Record, doc map[string]any) error {
	immutable := []string{"date", "utcOffset", "eventType", "device", "app", "srvCreated", "subject", "srvModified", "modifiedBy", "isValid"}
	for _, field := range immutable {
		if proposed, exists := doc[field]; exists {
			current := nsutils.PathValue(existing.ToMap(true), field)
			if field == "identifier" {
				continue
			}
			if fmt.Sprint(proposed) != fmt.Sprint(current) {
				return fmt.Errorf("Field %s cannot be modified by the client", field)
			}
		}
	}
	if identifier, ok := nsutils.StringField(doc, "identifier"); ok && identifier != "" && identifier != existing.Identifier() {
		delete(doc, "identifier")
	}
	return nil
}

func validatePatchDoc(existing nsutils.Record, doc map[string]any) error {
	immutable := []string{"identifier", "date", "utcOffset", "eventType", "device", "app", "srvCreated", "subject", "srvModified", "modifiedBy", "isValid"}
	for _, field := range immutable {
		if proposed, exists := doc[field]; exists {
			current := nsutils.PathValue(existing.ToMap(true), field)
			if fmt.Sprint(proposed) != fmt.Sprint(current) {
				return fmt.Errorf("Field %s cannot be modified by the client", field)
			}
		}
	}
	return nil
}

func normalizeV3Document(collection nsutils.Collection, doc map[string]any) (map[string]any, error) {
	clean, err := nsutils.NormalizeData(collection, doc)
	if err != nil {
		switch err.Error() {
		case "bad or missing date field":
			return nil, fmt.Errorf("Bad or missing date field")
		case "bad or missing utcOffset field":
			return nil, fmt.Errorf("Bad or missing utcOffset field")
		case "bad created_at field":
			return nil, fmt.Errorf("Bad created_at field")
		default:
			return nil, err
		}
	}

	if err := validateCreateDoc(collection, clean); err != nil {
		return nil, err
	}
	return clean, nil
}

func authErrorMessage(err error) string {
	switch {
	case err == nsutils.ErrBadToken, err == nsutils.ErrBadJWT:
		return "Bad access token or JWT"
	default:
		return "Missing or bad access token or JWT"
	}
}

func selectV3Fields(record nsutils.Record, fields []string) map[string]any {
	selected := nsutils.SelectFields(record, fields)
	delete(selected, "_id")
	return selected
}

func toInt64Compat(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	case float64:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, fmt.Errorf("not int")
	}
}
