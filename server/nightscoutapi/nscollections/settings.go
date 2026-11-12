package nscollections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
	"nighttroop.t1d.nz/utils/types"
)

type SettingsCollection struct{}

func NewSettingsCollection() SettingsCollection {
	return SettingsCollection{}
}

// setting is a setting as the Nightscout API shows it
type setting struct {
	Identifier string `json:"identifier"`
	Value      any    `json:"value"`
	Type       string `json:"type"`
}

// settingDocument is a setting as a document, changed and created when it was set
func settingDocument(source *settings.SettingAnyType) nsutils.Document {
	if source == nil {
		return stored[setting]{}
	}
	return stored[setting]{
		meta:  meta{id: source.Identifier, srvCreated: source.CreatedAt, srvModified: source.CreatedAt, occurredAt: source.CreatedAt, subject: source.Subject},
		model: setting{Identifier: source.Identifier, Value: source.Value, Type: source.TypeString},
	}
}

// ///
// Create Replace and Patch are pretty much the same things
// You have to pass an identifier along to match with the settings
// ///

func (c SettingsCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (nsutils.Document, bool, *nsutils.HTTPError) {
	var nsModel nightscoutmodels.Setting
	err := utils.ConvertViaJSON(data, &nsModel)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map data to model", "error", err, "collection", "settings")
		return nil, false, nsutils.BadDataBodyError
	}

	if nsModel.Identifier.String() == "" {
		slog.ErrorContext(ctx, "setting identifier not available", "data", data, "collection", "settings")
		return nil, false, nsutils.ErrNotFound
	}

	sm, herr := UpdateIndividualSetting(ctx, siteContext, nsModel.Identifier.String(), nsModel.Value)
	if herr != nil {
		slog.ErrorContext(ctx, "failed to update individual setting", "error", herr, "collection", "settings")
		return nil, false, herr
	}

	return settingDocument(sm), false, nil
}

func (c SettingsCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (nsutils.Document, bool, *nsutils.HTTPError) {
	if identifier == "" {
		slog.ErrorContext(ctx, "setting identifier not available", "data", patch, "collection", "settings")
		return nil, false, nsutils.ErrNotFound
	}
	// A document of an app's own, as AAPS's aaps, is kept as sent
	if !settings.IsSetting(identifier) {
		return putDocument(ctx, siteContext, identifier, patch, canCreate)
	}
	if patch["identifier"] == nil {
		patch["identifier"] = identifier
	}

	var nsModel nightscoutmodels.Setting
	err := utils.ConvertViaJSON(patch, &nsModel)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map data to model", "error", err, "collection", "settings")
		return nil, false, nsutils.BadDataBodyError
	}

	if nsModel.Identifier.String() == "" {
		nsModel.Identifier = ptr(types.IdentifierString(identifier))
	}

	if nsModel.Identifier.String() != identifier {
		slog.ErrorContext(ctx, "setting identifier mismatch", "data", patch, "collection", "settings")
		return nil, false, nsutils.ErrBadRequest
	}

	sm, herr := UpdateIndividualSetting(ctx, siteContext, nsModel.Identifier.String(), nsModel.Value)
	if herr != nil {
		slog.ErrorContext(ctx, "failed to update individual setting", "error", herr, "collection", "settings")
		return nil, false, herr
	}

	return settingDocument(sm), false, nil
}

func (c SettingsCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (nsutils.Document, *nsutils.HTTPError) {
	if identifier == "" {
		slog.ErrorContext(ctx, "setting identifier not available", "data", patch, "collection", "settings")
		return nil, nsutils.ErrNotFound
	}
	if !settings.IsSetting(identifier) {
		return nil, nsutils.ErrNotImplemented
	}
	if patch["identifier"] == nil {
		patch["identifier"] = identifier
	}

	record, _, err := c.Replace(ctx, siteContext, identifier, patch, false)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (c SettingsCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	if !settings.IsSetting(identifier) {
		return deleteDocument(ctx, siteContext, identifier)
	}
	_, herr := UpdateIndividualSetting(ctx, siteContext, identifier, nil)
	if herr != nil {
		slog.ErrorContext(ctx, "failed to update individual setting", "error", herr, "collection", "settings")
		return herr
	}

	return nil
}

func (c SettingsCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (nsutils.Document, *nsutils.HTTPError) {
	if !settings.IsSetting(identifier) {
		document, err := siteContext.SiteDB.Queries().GetSettingsDocument(ctx, sitedb.GetSettingsDocumentParams{ID: identifier})
		if err != nil {
			return nil, nsutils.ErrNotFound
		}
		return settingsDocument(document), nil
	}

	found, err := siteContext.GetOwnSettings().GetSettingAnyTypeByName(identifier)
	if err != nil {
		return nil, nsutils.ErrNotFound
	}

	return settingDocument(found), nil
}

func (c SettingsCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	// Apps' own documents, the last changed first, then Nighttroop's settings
	documents, err := siteContext.SiteDB.Queries().AllSettingsDocuments(ctx, sitedb.AllSettingsDocumentsParams{Limit: maxSettingsDocuments})
	if err != nil {
		return nil, nsutils.ErrInternal
	}
	docs := []nsutils.Document{}
	for _, document := range documents {
		docs = append(docs, settingsDocument(document))
	}
	for _, each := range siteContext.GetOwnSettings().GetAllSettingAnyTypes() {
		docs = append(docs, settingDocument(each))
	}
	return docs, nil
}

// History is every setting, as their changes are not kept
func (c SettingsCollection) History(ctx context.Context, siteContext *httputils.SiteContext, since int64, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	return c.Search(ctx, siteContext, query)
}

func (c SettingsCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c SettingsCollection) DeleteQuery(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func UpdateIndividualSetting(ctx context.Context, siteContext *httputils.SiteContext, identifier string, value any) (*settings.SettingAnyType, *nsutils.HTTPError) {
	settingsMap := siteContext.GetOwnSettings()
	identitySettingsMap := settingsMap.IdentitySettingsMap

	err := identitySettingsMap.SetSettingByName(identifier, value, &siteContext.IdentityModel.CreatedAt, ptr("identity"))
	if err != nil {
		slog.ErrorContext(ctx, "failed to set setting in identity settings map", "error", err, "collection", "settings")
		return nil, nsutils.BadDataBodyErrorWithHint(err.Error())
	}
	if err := settingsMap.Consistent(); err != nil {
		return nil, nsutils.BadDataBodyErrorWithHint(err.Error())
	}

	params := &controldb.UpdateSiteIdentitySettingsParams{
		ID:     siteContext.IdentityModel.ID,
		SiteID: siteContext.IdentityModel.SiteID,
	}
	settings.SetSettings(params, identitySettingsMap)

	err = siteContext.ControlDB.Queries.UpdateSiteIdentitySettings(ctx, *params)
	if err != nil {
		if nsErr := nsutils.GetSQLTranslatedError(err); nsErr != nil {
			return nil, nsErr
		}
		slog.ErrorContext(ctx, "failed to update identity settings", "error", err, "collection", "settings")
		return nil, nsutils.ErrInternal
	}

	return identitySettingsMap.GetSettingAnyTypeByName(identifier), nil
}
