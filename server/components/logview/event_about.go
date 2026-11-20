package logview

import (
	"context"
	"encoding/json"
	"strings"

	"nighttroop.t1d.nz/utils/i18n"
)

// EventAbout is what an event's attributes say about it in a few words, to tell events apart in a list without
// opening each: who it was for and who did it, the connector, how many records, why, and where a sign in was from.
// "" when its attributes say none of these.
func EventAbout(ctx context.Context, raw json.RawMessage) string {
	var about struct {
		IdentityName  string `json:"identity_name"`
		By            string `json:"by"`
		ConnectorType string `json:"connector_type"`
		Records       *int   `json:"records"`
		Reason        string `json:"reason"`
		// From is a sign in's place, and a backfill's start or a change's old value otherwise, which are left out
		From json.RawMessage `json:"from"`
	}
	_ = json.Unmarshal(raw, &about)
	var place struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}
	_ = json.Unmarshal(about.From, &place)

	parts := []string{about.IdentityName}
	if about.By != "" && about.By != about.IdentityName {
		parts = append(parts, i18n.T(ctx, "common.log_view.about_by", i18n.Data{"By": about.By}))
	}
	parts = append(parts, about.ConnectorType)
	if about.Records != nil {
		parts = append(parts, i18n.T(ctx, "common.log_view.about_records", i18n.Data{"Records": *about.Records}))
	}
	parts = append(parts, about.Reason, strings.Trim(place.City+", "+place.Country, ", "))
	var said []string
	for _, part := range parts {
		if part != "" {
			said = append(said, part)
		}
	}
	return strings.Join(said, " · ")
}
