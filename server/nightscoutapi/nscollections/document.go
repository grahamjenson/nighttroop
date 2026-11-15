package nscollections

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

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
// Only the values not already written in that form are read again.
func (s stored[NS]) JSON(v3 bool, only []string) json.RawMessage {
	var model any = &s.model
	if forV3, ok := any(s.model).(interface{ V3() any }); ok && v3 {
		model = forV3.V3()
	}
	raw, _ := json.Marshal(model)
	server, without := s.server(v3)
	// The fields are kept on the stack, to be written once they are sorted
	var room [32]field
	fields := room[:0]
	add := func(name string, value []byte) {
		if name != without && (len(only) == 0 || (name != "_id" && slices.Contains(only, name))) {
			fields = append(fields, field{name, value})
		}
	}
	// The names are cut from one copy of the model's JSON, so that each is not copied
	text := string(raw)
	decoder := jsontext.NewDecoder(bytes.NewBuffer(raw), jsontext.AllowDuplicateNames(true))
	if token, err := decoder.ReadToken(); err == nil && token.Kind() == '{' {
		for decoder.PeekKind() == '"' {
			quoted, _ := decoder.ReadValue()
			name := text[int(decoder.InputOffset())-len(quoted)+1 : decoder.InputOffset()-1]
			if bytes.IndexByte(quoted, '\\') >= 0 {
				unquoted, _ := jsontext.AppendUnquote(nil, quoted)
				name = string(unquoted)
			}
			// The value as it is in the model's JSON, which the decoder's own is only until it reads on
			decoded, _ := decoder.ReadValue()
			value := raw[int(decoder.InputOffset())-len(decoded) : decoder.InputOffset()]
			if name == "units" {
				var units string
				if json.Unmarshal(value, &units) == nil && s.unitsLabels[units] != "" {
					value = quote(s.unitsLabels[units])
				}
			}
			if !slices.ContainsFunc(server, func(f field) bool { return f.name == name }) {
				add(name, value)
			}
		}
	}
	for _, f := range server {
		add(f.name, f.value)
	}
	// Names in the order Canonicalize sorts them, which is the order of their bytes when they are all ASCII
	ascii := !slices.ContainsFunc(fields, func(f field) bool {
		return strings.ContainsFunc(f.name, func(r rune) bool { return r >= utf8.RuneSelf })
	})
	slices.SortFunc(fields, func(a, b field) int { return strings.Compare(a.name, b.name) })
	out := append(make([]byte, 0, len(raw)+256), '{')
	for i, f := range fields {
		if i > 0 {
			out = append(out, ',')
		}
		out, _ = jsontext.AppendQuote(out, f.name)
		out = append(out, ':')
		if !ascii || canonical(f.value) {
			out = append(out, f.value...)
			continue
		}
		value := jsontext.Value(slices.Clone(f.value))
		_ = value.Canonicalize(jsontext.EscapeForJS(true))
		out = append(out, value...)
	}
	written := jsontext.Value(append(out, '}'))
	if !ascii {
		_ = written.Canonicalize(jsontext.EscapeForJS(true))
	}
	return json.RawMessage(written)
}

// canonical is whether a value is already as Canonicalize writes it: a literal, an ASCII string with no escapes, a
// whole number short enough to be exact as a float64, or a number as its float64 is written in fixed notation
func canonical(value []byte) bool {
	switch value[0] {
	case 't', 'f', 'n':
		return true
	case '"':
		return !slices.ContainsFunc(value, func(c byte) bool { return c == '\\' || c >= utf8.RuneSelf })
	case '{', '[':
		return false
	}
	if len(value) <= 15 && !bytes.ContainsAny(value, ".eE") && string(value) != "-0" {
		return true
	}
	number, err := strconv.ParseFloat(string(value), 64)
	var formatted [32]byte
	return err == nil && (string(value) == "0" || math.Abs(number) >= 1e-6 && math.Abs(number) < 1e21) &&
		bytes.Equal(strconv.AppendFloat(formatted[:0], number, 'f', -1, 64), value)
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
	// Room for the server's fields after the aliases, so it is made once
	fields := make([]field, 0, len(aliases)+7)
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
