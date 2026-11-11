package sitemodels

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/display"
)

// SiteLoopSettings are a set of a loop app's settings, kept once however often they were sent, with when they were
// first and last seen. ID is the SHA-256 of the app and the settings.
type SiteLoopSettings struct {
	ID                  string          `json:"id"`
	App                 string          `json:"app"`
	Source              string          `json:"source"`
	Settings            json.RawMessage `json:"settings"`
	FirstSeenAt         int64           `json:"first_seen_at"`
	LastSeenAt          int64           `json:"last_seen_at"`
	CreatedByDeviceName *string         `json:"created_by_device_name"`
	CreatedAt           int64           `json:"created_at"`
	// PreviousID is the set of the same app last seen before this one, and Changes are the settings that differ from
	// it, glucose in Units; they are only on a set read alone, and an app's first set has none
	PreviousID *string               `json:"previous_id"`
	Changes    []*LoopSettingsChange `json:"changes"`
	Units      *consts.Units         `json:"units"`
}

// LoopSettingsChange is a setting that differs from the set before: where it is in the settings, its names joined by
// dots, with a list of presets' by each one's name, and its value before and after, nil when it was not sent
type LoopSettingsChange struct {
	Field string  `json:"field"`
	Old   *string `json:"old"`
	New   *string `json:"new"`
}

// LoopSettingsFromDB is the site model of a loop_settings row
func LoopSettingsFromDB(row *sitedb.LoopSetting) *SiteLoopSettings {
	if row == nil {
		return nil
	}
	return &SiteLoopSettings{ID: row.ID, App: row.App, Source: row.Source, Settings: json.RawMessage(row.SettingsJson),
		FirstSeenAt: row.FirstSeenAt, LastSeenAt: row.LastSeenAt, CreatedByDeviceName: row.CreatedByDeviceName, CreatedAt: row.CreatedAt}
}

// loopGlucoseSettings are the settings Loop and Trio send glucose in, Loop's in its profile's units and Trio's in
// mg/dL; AAPS's configuration has none
var loopGlucoseSettings = map[string]bool{"minimumBGGuard": true, "preMealTargetRange": true, "targetRange": true, "target": true}

// LoopSettingsChanges are the settings that differ between two of an app's sets, by field, with glucose in units
func LoopSettingsChanges(app string, before, after json.RawMessage, units *consts.Units) []*LoopSettingsChange {
	old, changed := map[string]string{}, map[string]string{}
	flattenLoopSettings(app != "aaps", "", decodeLoopSettings(before), units, old)
	flattenLoopSettings(app != "aaps", "", decodeLoopSettings(after), units, changed)
	fields := maps.Clone(old)
	maps.Copy(fields, changed)
	changes := []*LoopSettingsChange{}
	for _, field := range slices.Sorted(maps.Keys(fields)) {
		was, wasSent := old[field]
		is, isSent := changed[field]
		if wasSent != isSent || was != is {
			change := &LoopSettingsChange{Field: field}
			if wasSent {
				change.Old = &was
			}
			if isSent {
				change.New = &is
			}
			changes = append(changes, change)
		}
	}
	return changes
}

// decodeLoopSettings are the settings as sent, or none when they cannot be read
func decodeLoopSettings(raw json.RawMessage) map[string]any {
	settings := map[string]any{}
	_ = json.Unmarshal(raw, &settings)
	return settings
}

// flattenLoopSettings puts each setting under path in out, as text, by its field: an object's by each of its names,
// and a list of presets' by each one's name, or its place when it has none, so adding one does not change the rest
func flattenLoopSettings(glucose bool, path string, value any, units *consts.Units, out map[string]string) {
	field := func(name string) string {
		if path == "" {
			return name
		}
		return path + "." + name
	}
	switch v := value.(type) {
	case map[string]any:
		for name, child := range v {
			flattenLoopSettings(glucose, field(name), child, units, out)
		}
		return
	case []any:
		if slices.ContainsFunc(v, func(item any) bool { _, ok := item.(map[string]any); return ok }) {
			for i, item := range v {
				name := strconv.Itoa(i + 1)
				// A preset is put under its name unless one before it had the same name
				if preset, ok := item.(map[string]any); ok {
					if presetName, ok := preset["name"].(string); ok && presetName != "" && out[field(presetName)+".name"] == "" {
						name = presetName
					}
				}
				flattenLoopSettings(glucose, field(name), item, units, out)
			}
			return
		}
	}
	out[path] = loopSettingText(glucose && loopGlucoseSettings[path[strings.LastIndex(path, ".")+1:]], value, units)
}

// loopSettingText is a setting as it reads: a glucose, or a range of them, in units, text as it is, and anything else
// as JSON
func loopSettingText(glucose bool, value any, units *consts.Units) string {
	values, isList := value.([]any)
	if !isList {
		values = []any{value}
	}
	texts := []string{}
	for _, v := range values {
		mgdl, isNumber := v.(float64)
		if !glucose || !isNumber {
			texts = nil
			break
		}
		// Loop sends glucose in its profile's units, which are not kept with it, so they are guessed from its size
		var sentIn *consts.Units
		texts = append(texts, display.Glucose(units.FromMGDL(sentIn.ToMGDL(mgdl)), units))
	}
	if len(texts) > 0 {
		return strings.TrimSpace(strings.Join(texts, "–") + " " + display.Units(units))
	}
	if text, ok := value.(string); ok {
		return text
	}
	text, _ := json.Marshal(value)
	return string(text)
}
