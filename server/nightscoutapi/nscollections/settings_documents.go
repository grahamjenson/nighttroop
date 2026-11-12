package nscollections

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxSettingsDocuments is how many a site keeps: AAPS's aaps, aaps-state and clientcontrol-* ones, but not a DB full
const maxSettingsDocuments = 16

var ErrTooManySettingsDocuments = nsutils.BadDataBodyErrorWithHint("a site keeps at most 16 settings documents of apps' own")

// settingsDocument is an app's settings document as it was sent, with the server's fields
func settingsDocument(dbModel *sitedb.SettingsDocument) nsutils.Document {
	m := meta{id: dbModel.ID, srvCreated: dbModel.CreatedAt, srvModified: dbModel.UpdatedAt, occurredAt: dbModel.OccurredAt,
		utcOffset: dbModel.UtcOffsetMinutes, subject: dbModel.CreatedByIdentityName}
	return stored[json.RawMessage]{meta: m, model: json.RawMessage(dbModel.DocumentJson)}
}

// Record is an app's settings document for live updates, and false for any other model
func (c SettingsCollection) Record(model any, siteContext *httputils.SiteContext) (nsutils.Document, bool) {
	if dbModel, ok := model.(*sitedb.SettingsDocument); ok && dbModel != nil {
		return settingsDocument(dbModel), true
	}
	return nil, false
}

// putDocument keeps an app's settings document as sent, replacing one, as API v3's PUT does, and whether it created it
func putDocument(ctx context.Context, siteContext *httputils.SiteContext, identifier string, doc map[string]any, canCreate bool) (nsutils.Document, bool, *nsutils.HTTPError) {
	var dated struct{ Date, UtcOffset *int64 }
	documentJSON, err := json.Marshal(doc)
	if err != nil || json.Unmarshal(documentJSON, &dated) != nil || dated.Date == nil {
		return nil, false, nsutils.BadDataBodyErrorWithHint("a settings document needs a date in milliseconds")
	}
	q := siteContext.SiteDB.Queries()
	if _, err := q.GetSettingsDocument(ctx, sitedb.GetSettingsDocumentParams{ID: identifier}); err != nil && !canCreate {
		return nil, false, nsutils.ErrMissingPermission("api:settings:create")
	}
	dbModel, err := q.PutSettingsDocument(ctx, sitedb.PutSettingsDocumentParams{ID: identifier, DocumentJson: string(documentJSON),
		OccurredAt: *dated.Date, UtcOffsetMinutes: *cmp.Or(dated.UtcOffset, new(int64)), CreatedByIdentityName: siteContext.IdentityModel.Name, Max: maxSettingsDocuments})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, ErrTooManySettingsDocuments
	} else if err != nil {
		return nil, false, cmp.Or(nsutils.GetSQLTranslatedError(err), nsutils.ErrInternal)
	}
	created, action := dbModel.CreatedAt == dbModel.UpdatedAt, events.Updated
	if created {
		action = events.Created
	}
	siteContext.PublishModelChange(action, dbModel)
	return settingsDocument(dbModel), created, nil
}

// deleteDocument deletes an app's settings document for good, as presets are, so it no longer counts to the site's
func deleteDocument(ctx context.Context, siteContext *httputils.SiteContext, identifier string) *nsutils.HTTPError {
	dbModel, err := siteContext.SiteDB.Queries().DeleteSettingsDocument(ctx, sitedb.DeleteSettingsDocumentParams{ID: identifier})
	if err != nil {
		return nsutils.ErrNotFound
	}
	siteContext.PublishModelChange(events.Deleted, dbModel)
	return nil
}
