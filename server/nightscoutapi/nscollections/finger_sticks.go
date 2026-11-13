package nscollections

import (
	"cmp"
	"encoding/json"
	"maps"
	"slices"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// AAPS takes a finger stick only from a BG Check treatment, and xDrip+ sends one only as a meter reading (mbg), so API
// v3 gives AAPS, with the treatments, each meter reading in no treatment as a read-only BG Check. A careportal BG Check
// has its reading, so it is not given twice.

// FingerSuffix ends the identifier of a meter reading given as a BG Check: the reading's identifier with it
const FingerSuffix = "finger"

// FingerStick is whether a reading is one API v3 can give as a BG Check, when no treatment has it: a meter reading not
// from AAPS, which has its own
func FingerStick(reading *sitedb.GlucoseReading) bool {
	return reading.ReadingType == "mbg" && (reading.CreatedByAppName == nil || *reading.CreatedByAppName != "AAPS")
}

// fieldsDocument is a document made of its fields
type fieldsDocument struct {
	id                   string
	modified, occurredAt int64
	fields               map[string]any
}

func (d fieldsDocument) Identifier() string         { return d.id }
func (d fieldsDocument) Modified() int64            { return d.modified }
func (d fieldsDocument) OccurredAt() int64          { return d.occurredAt }
func (d fieldsDocument) Fields(bool) map[string]any { return maps.Clone(d.fields) }
func (d fieldsDocument) JSON(_ bool, only []string) json.RawMessage {
	fields := d.Fields(true)
	if len(only) > 0 {
		maps.DeleteFunc(fields, func(name string, _ any) bool { return !slices.Contains(only, name) })
	}
	raw, _ := json.Marshal(fields)
	return raw
}

// BGCheck is an entry, a meter reading, as a read-only BG Check, as API v3 gives it, and false for any other entry
func BGCheck(doc nsutils.Document, siteContext *httputils.SiteContext) (nsutils.Document, bool) {
	entry := doc.Fields(true)
	id := doc.Identifier() + FingerSuffix
	if entry["type"] != "mbg" || nsutils.ValidateIdentifierIsntBad(id) != nil {
		return nil, false
	}
	fields := map[string]any{"identifier": id, "eventType": "BG Check", "glucose": entry["mbg"], "glucoseType": "Finger",
		"units": siteContext.GetSettings().GetUnitsLabel(consts.UnitsMGDL), "enteredBy": cmp.Or(entry["device"], entry["subject"]),
		"isReadOnly": true}
	for _, name := range []string{"date", "created_at", "utcOffset", "srvCreated", "srvModified", "subject", "isValid", "app"} {
		if value, ok := entry[name]; ok {
			fields[name] = value
		}
	}
	return fieldsDocument{id: id, modified: doc.Modified(), occurredAt: doc.OccurredAt(), fields: fields}, true
}

