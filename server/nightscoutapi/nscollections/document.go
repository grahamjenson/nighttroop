package nscollections

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"slices"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

// stored is a collection's Nightscout model, with the server's fields from its row. It is only made into JSON fields
// when it is written.
type stored[NS any] struct {
	meta
	model NS
	// unitsLabels are how the identity's settings write units, by Nighttroop's, when they differ
	unitsLabels map[string]string
}

// meta are the fields the server keeps for every document
type meta struct {
	id                      string
	srvCreated, srvModified int64
	occurredAt, utcOffset   int64
	subject                 string
	deletedAt               *int64
}

func (m meta) Identifier() string { return m.id }
func (m meta) Modified() int64    { return m.srvModified }
func (m meta) OccurredAt() int64  { return m.occurredAt }

// row is a database model as a document's server fields
func row(dbModel SharedDBFields) meta {
	m := meta{id: dbModel.GetID(), srvCreated: dbModel.GetCreatedAt(), srvModified: dbModel.GetUpdatedAt(),
		occurredAt: dbModel.GetOccurredAt(), subject: dbModel.GetCreatedByIdentityName(), deletedAt: dbModel.GetDeletedAt()}
	if offset := dbModel.GetUtcOffsetMinutes(); offset != nil {
		m.utcOffset = *offset
	}
	return m
}

// Fields is it as JSON fields, each value as JSON reads it
func (s stored[NS]) Fields(v3 bool) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal(s.JSON(v3, nil), &out)
	return out
}

// field is a JSON object's field, with its value as JSON
type field struct {
	name  string
	value []byte
}

// JSON is it as a JSON object of its fields, or of only the given fields less _id when there are any. As a search
// writes many, the model's fields are copied from its JSON rather than read back into values, and the object is then
// written as its fields read back would be: names in order, numbers as JSON reads them, strings as the API escapes them.
func (s stored[NS]) JSON(v3 bool, only []string) json.RawMessage {
	var model any = &s.model
	if forV3, ok := any(s.model).(interface{ V3() any }); ok && v3 {
		model = forV3.V3()
	}
	raw, _ := json.Marshal(model)
	server, without := s.server(v3)
	out := append(make([]byte, 0, len(raw)+256), '{')
	add := func(name string, value []byte) {
		if name == without || (len(only) > 0 && (name == "_id" || !slices.Contains(only, name))) {
			return
		}
		if len(out) > 1 {
			out = append(out, ',')
		}
		out, _ = jsontext.AppendQuote(out, name)
		out = append(append(out, ':'), value...)
	}
	decoder := jsontext.NewDecoder(bytes.NewBuffer(raw), jsontext.AllowDuplicateNames(true))
	if token, err := decoder.ReadToken(); err == nil && token.Kind() == '{' {
		for decoder.PeekKind() == '"' {
			token, _ := decoder.ReadToken()
			name := token.String()
			// The value as it is in the model's JSON, which the decoder's own is only until it reads on
			decoded, _ := decoder.ReadValue()
			value := raw[int(decoder.InputOffset())-len(decoded) : decoder.InputOffset()]
			var units string
			if name == "units" && json.Unmarshal(value, &units) == nil && s.unitsLabels[units] != "" {
				value = quote(s.unitsLabels[units])
			}
			if !slices.ContainsFunc(server, func(f field) bool { return f.name == name }) {
				add(name, value)
			}
		}
	}
	for _, f := range server {
		add(f.name, f.value)
	}
	written := jsontext.Value(append(out, '}'))
	_ = written.Canonicalize(jsontext.EscapeForJS(true))
	return json.RawMessage(written)
}

// carbsSuffix ends the identifier of a treatment's carbs part, its treatment's identifier with it
const carbsSuffix = "carbs"

// CarbsPartOf is the identifier of the treatment a carbs part's identifier is of
func CarbsPartOf(identifier string) (string, bool) {
	return strings.CutSuffix(identifier, carbsSuffix)
}

// V3Parts are the documents API v3 lists and its storage socket sends for it: a treatment with both insulin and carbs
// as its insulin and its carbs, as AAPS reads only one of them from a document, the carbs with the treatment's
// identifier and carbsSuffix. One whose identifier is too long to add carbsSuffix to stays one.
func (s stored[NS]) V3Parts() []nsutils.Document {
	treatment, ok := any(s.model).(nightscoutmodels.Treatment)
	if !ok || nsutils.ValidateIdentifierIsntBad(s.id+carbsSuffix) != nil {
		return []nsutils.Document{s}
	}
	insulin, carbs, split := treatment.SplitCarbs()
	if !split {
		return []nsutils.Document{s}
	}
	insulinPart := stored[nightscoutmodels.Treatment]{meta: s.meta, model: insulin, unitsLabels: s.unitsLabels}
	carbsPart := stored[nightscoutmodels.Treatment]{meta: s.meta, model: carbs, unitsLabels: s.unitsLabels}
	carbsPart.id += carbsSuffix
	return []nsutils.Document{insulinPart, carbsPart}
}

// server are the server's fields, over the model's, and the one of _id and identifier the API version does not have.
// Every document is written with a created_at, and a date on API v3, and with its collection's time aliases.
func (s stored[NS]) server(v3 bool) ([]field, string) {
	aliases := types.TimeAliases{{Name: "created_at"}}
	if v3 {
		aliases = append(aliases, types.TimeAlias{Name: "date", Millis: true})
	}
	if timed, ok := any(s.model).(interface{ TimeAliases() types.TimeAliases }); ok {
		aliases = append(aliases, timed.TimeAliases()...)
	}
	fields := []field{}
	for _, written := range aliases.Written(s.occurredAt, v3) {
		if !slices.ContainsFunc(fields, func(f field) bool { return f.name == written[0] }) {
			fields = append(fields, field{written[0], []byte(written[1])})
		}
	}
	fields = append(fields, field{"utcOffset", strconv.AppendInt(nil, s.utcOffset, 10)})
	if !v3 {
		return append(fields, field{"_id", quote(s.id)}), "identifier"
	}
	return append(fields, field{"identifier", quote(s.id)},
		field{"srvCreated", strconv.AppendInt(nil, s.srvCreated, 10)}, field{"srvModified", strconv.AppendInt(nil, s.srvModified, 10)},
		field{"subject", quote(s.subject)}, field{"isValid", strconv.AppendBool(nil, s.deletedAt == nil)},
		field{"isReadOnly", []byte("false")}), "_id"
}

// quote is a string as JSON
func quote(s string) []byte {
	out, _ := jsontext.AppendQuote(nil, s)
	return out
}

// unitsLabels are how the identity's settings write each units in the Nightscout API, where it is not as Nighttroop
// writes them, as xDrip4iOS reads a BG check's glucose as mg/dL only when its units are mg/dl
func unitsLabels(siteContext *httputils.SiteContext) map[string]string {
	if siteContext == nil {
		return nil
	}
	labels := map[string]string{}
	for _, units := range []consts.Units{consts.UnitsMGDL, consts.UnitsMMOLL} {
		if label := siteContext.GetSettings().GetUnitsLabel(units); label != string(units) {
			labels[string(units)] = label
		}
	}
	return labels
}
