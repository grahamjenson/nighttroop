package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"maps"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// The settings collection's keys whose documents Nighttroop keeps for an app as they are sent, shared by the site's
// identities, rather than as a setting of its own: AAPS's running config, and the messages AAPS and AAPSClient
// send each other to pair and control it
const (
	generalAAPS          = "aaps"
	generalClientControl = "aaps_clientcontrol"
	// maxGeneralSettings is the most a site keeps, and maxGeneralSettingBytes the largest one's JSON
	maxGeneralSettings     = 100
	maxGeneralSettingBytes = 256 * 1024
	// staleClientControl is how long a client control message is kept unchanged before a new one may take its place
	staleClientControl = 7 * 24 * time.Hour
)

// generalSettingType is the kind of general setting a key is, or "" for the site's own settings
func generalSettingType(key string) string {
	switch {
	case key == "aaps" || key == "aaps-state":
		return generalAAPS
	case strings.HasPrefix(key, generalClientControl+"_"):
		return generalClientControl
	}
	return ""
}

// generalSettingDocument is a kept document as the API shows it: at the date it gives, by whoever first sent it
func generalSettingDocument(row *sitedb.GeneralSetting) nsutils.Document {
	var sent struct {
		Date      *int64 `json:"date"`
		UtcOffset *int64 `json:"utcOffset"`
	}
	_ = json.Unmarshal([]byte(row.ValueJson), &sent)
	m := meta{id: row.Key, srvCreated: row.CreatedAt, srvModified: row.UpdatedAt, occurredAt: row.UpdatedAt, subject: row.CreatedByIdentityName}
	if sent.Date != nil {
		m.occurredAt = *sent.Date
	}
	if sent.UtcOffset != nil {
		m.utcOffset = *sent.UtcOffset
	}
	return stored[json.RawMessage]{meta: m, model: json.RawMessage(row.ValueJson)}
}

// saveGeneralSetting keeps the document under its key, replacing the one there, or with patch adding its fields to
// it. A new key is refused once the site has its most, after making room by deleting client control messages left
// unchanged a week. It says whether the key is new.
func saveGeneralSetting(ctx context.Context, siteContext *httputils.SiteContext, key string, doc map[string]any, patch bool) (nsutils.Document, bool, *nsutils.HTTPError) {
	q := siteContext.SiteDB.Queries()
	existing, err := q.GetGeneralSetting(ctx, sitedb.GetGeneralSettingParams{Key: key})
	created := errors.Is(err, sql.ErrNoRows)
	switch {
	case created && patch:
		return nil, false, nsutils.ErrNotFound
	case created:
		if err := q.DeleteGeneralSettingsBefore(ctx, sitedb.DeleteGeneralSettingsBeforeParams{Type: generalClientControl,
			UpdatedAt: time.Now().Add(-staleClientControl).UnixMilli()}); err != nil {
			return nil, false, nsutils.ErrInternal
		}
		if count, err := q.CountGeneralSettings(ctx); err != nil || count >= maxGeneralSettings {
			slog.WarnContext(ctx, "general settings full", "error", err, "key", key)
			return nil, false, nsutils.BadDataBodyErrorWithHint("too many settings documents")
		}
	case err != nil:
		return nil, false, nsutils.ErrInternal
	case patch:
		merged := map[string]any{}
		_ = json.Unmarshal([]byte(existing.ValueJson), &merged)
		maps.Copy(merged, doc)
		doc = merged
	}
	// The server's own fields are written by the server, not kept from what was sent
	for _, field := range []string{"_id", "srvCreated", "srvModified", "subject", "isValid", "isReadOnly"} {
		delete(doc, field)
	}
	doc["identifier"] = key
	value, err := json.Marshal(doc)
	if err != nil || len(value) > maxGeneralSettingBytes {
		slog.WarnContext(ctx, "general setting too large", "key", key, "bytes", len(value))
		return nil, false, nsutils.BadDataBodyErrorWithHint("settings document too large")
	}
	row, err := q.UpsertGeneralSetting(ctx, sitedb.UpsertGeneralSettingParams{Key: key, Type: generalSettingType(key),
		ValueJson: string(value), CreatedByIdentityName: siteContext.IdentityModel.Name})
	if err != nil {
		slog.ErrorContext(ctx, "failed to save general setting", "error", err, "key", key)
		return nil, false, nsutils.ErrInternal
	}
	action := events.Updated
	if created {
		action = events.Created
	}
	siteContext.PublishModelChange(action, row)
	return generalSettingDocument(row), created, nil
}

// getGeneralSetting is the document kept under the key
func getGeneralSetting(ctx context.Context, siteContext *httputils.SiteContext, key string) (nsutils.Document, *nsutils.HTTPError) {
	row, err := siteContext.SiteDB.Queries().GetGeneralSetting(ctx, sitedb.GetGeneralSettingParams{Key: key})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nsutils.ErrNotFound
	} else if err != nil {
		return nil, nsutils.ErrInternal
	}
	return generalSettingDocument(row), nil
}

// generalSettingsSince are the documents changed since the time, newest first
func generalSettingsSince(ctx context.Context, siteContext *httputils.SiteContext, since int64) ([]nsutils.Document, *nsutils.HTTPError) {
	rows, err := siteContext.SiteDB.Queries().SearchGeneralSettings(ctx, sitedb.SearchGeneralSettingsParams{UpdatedAtGTE: since, Limit: maxGeneralSettings})
	if err != nil {
		return nil, nsutils.ErrInternal
	}
	docs := make([]nsutils.Document, len(rows))
	for i, row := range rows {
		docs[i] = generalSettingDocument(row)
	}
	return docs, nil
}

// deleteGeneralSetting deletes the document kept under the key
func deleteGeneralSetting(ctx context.Context, siteContext *httputils.SiteContext, key string) *nsutils.HTTPError {
	row, err := siteContext.SiteDB.Queries().DeleteGeneralSetting(ctx, sitedb.DeleteGeneralSettingParams{Key: key})
	if errors.Is(err, sql.ErrNoRows) {
		return nsutils.ErrNotFound
	} else if err != nil {
		return nsutils.ErrInternal
	}
	siteContext.PublishModelChange(events.Deleted, row)
	return nil
}

// Record is a kept document that changed, for the sockets to send
func (c SettingsCollection) Record(model any, siteContext *httputils.SiteContext) (nsutils.Document, bool) {
	row, ok := model.(*sitedb.GeneralSetting)
	if !ok {
		return nil, false
	}
	return generalSettingDocument(row), true
}
