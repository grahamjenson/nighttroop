package nscollections

import (
	"encoding/json"
	"maps"
	"strconv"

	"nighttroop.t1d.nz/utils/consts"
)

// mergeFields is the stored document with the fields sent laid over it as a JSON merge patch (RFC 7386) lays them: an
// object sent is merged into the stored one field by field, so a device status's pump sent with only its reservoir
// keeps its model, a field sent as null is taken out, and any other value sent, a list too, takes the place of the
// stored one. A field sent also takes the place of the stored ones that say the same (aliases), as a treatment's
// duration does its durationInMilliseconds. When glucose or its units are sent, the stored glucose not sent (the
// model's glucose fields) is given in the units sent, or guessed from the glucose sent.
func mergeFields(stored, sent map[string]any, aliases map[string][]string, glucose []string) map[string]any {
	merged := maps.Clone(stored)
	for name := range sent {
		for _, alias := range aliases[name] {
			delete(merged, alias)
		}
	}
	if units, ok := sentUnits(sent, glucose); ok {
		from := documentUnits(stored)
		for _, name := range glucose {
			if value, ok := merged[name]; ok {
				merged[name] = convertGlucose(value, from, units)
			}
		}
		merged["units"] = string(units)
	}
	return mergePatch(merged, sent)
}

// mergePatch is the patch's fields laid over the object's, objects merged into objects, see mergeFields
func mergePatch(object, patch map[string]any) map[string]any {
	merged := maps.Clone(object)
	if merged == nil {
		merged = map[string]any{}
	}
	for name, value := range patch {
		sentObject, sentIsObject := value.(map[string]any)
		storedObject, storedIsObject := merged[name].(map[string]any)
		switch {
		case value == nil:
			delete(merged, name)
		case sentIsObject && storedIsObject:
			merged[name] = mergePatch(storedObject, sentObject)
		default:
			merged[name] = value
		}
	}
	return merged
}

// sentUnits are the units of the glucose sent: its units, or those its first glucose is guessed to be in, as a create
// reads glucose sent without units
func sentUnits(sent map[string]any, glucose []string) (consts.Units, bool) {
	if units, ok := sent["units"].(string); ok {
		parsed, err := consts.ParseUnits(units)
		return parsed, err == nil
	}
	for _, name := range glucose {
		if value, ok := firstNumber(sent[name]); ok {
			return consts.GuessUnits(value), true
		}
	}
	return "", false
}

// documentUnits are the units of a stored document's glucose, nil when it has none, so each value's are guessed
func documentUnits(stored map[string]any) *consts.Units {
	if units, ok := stored["units"].(string); ok {
		if parsed, err := consts.ParseUnits(units); err == nil {
			return &parsed
		}
	}
	return nil
}

// convertGlucose is a glucose value, or a list of them like Loop's correctionRange, in units from given in units to
func convertGlucose(value any, from *consts.Units, to consts.Units) any {
	if list, ok := value.([]any); ok {
		converted := make([]any, len(list))
		for i, item := range list {
			converted[i] = convertGlucose(item, from, to)
		}
		return converted
	}
	number, ok := firstNumber(value)
	if !ok {
		return value
	}
	return to.FromMGDL(from.ToMGDL(number))
}

// firstNumber is a JSON value as a number, a string of one too, or a list's first one
func firstNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case json.Number:
		number, err := v.Float64()
		return number, err == nil
	case string:
		number, err := strconv.ParseFloat(v, 64)
		return number, err == nil
	case []any:
		if len(v) > 0 {
			return firstNumber(v[0])
		}
	}
	return 0, false
}
