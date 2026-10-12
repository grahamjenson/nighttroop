package collections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
)

type SettingsCollection struct {
	converter func(source *settings.SettingAnyType, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewSettingsCollection() SettingsCollection {
	return SettingsCollection{
		converter: ConvertFromAnyTypeToSetting,
	}
}

func ConvertFromAnyTypeToSetting(source *settings.SettingAnyType, context *httputils.SiteContext, raw bool) nsutils.Record {
	if source == nil {
		return nsutils.Record{}
	}

	record := nsutils.Record{
		Collection: string("settings"),
		ID:         source.Identifier,
		Data: map[string]any{
			"identifier": source.Identifier,
			"value":      source.Value,
			"type":       source.TypeString,
		},
		SrvCreated:  source.CreatedAt,
		SrvModified: source.CreatedAt,
		Subject:     source.Subject,
		UTCOffset:   0,
		OccuredAt:   source.CreatedAt,
	}
	return record
}

// ///
// Create Replace and Patch are pretty much the same things
// You have to pass an identifier along to match with the settings
// ///

func (c SettingsCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	var nsModel nightscout.Setting
	err := utils.FromMap(data, &nsModel)
	if err != nil {
		slog.Error("failed to map data to model", "error", err, "collection", "settings")
		return nil, false, nsutils.BadDataBodyError
	}

	if nsModel.Identifier.String() == "" {
		slog.Error("setting identifier not available", "data", data, "collection", "settings")
		return nil, false, nsutils.ErrNotFound
	}

	sm, herr := UpdateIndividualSetting(ctx, siteContext, nsModel.Identifier.String(), nsModel.Value)
	if herr != nil {
		slog.Error("failed to update individual setting", "error", herr, "collection", "settings")
		return nil, false, herr
	}

	return ptr(c.converter(sm, siteContext, false)), false, nil
}

func (c SettingsCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	if identifier == "" {
		slog.Error("setting identifier not available", "data", patch, "collection", "settings")
		return nil, false, nsutils.ErrNotFound
	}
	if patch["identifier"] == nil {
		patch["identifier"] = identifier
	}

	var nsModel nightscout.Setting
	err := utils.FromMap(patch, &nsModel)
	if err != nil {
		slog.Error("failed to map data to model", "error", err, "collection", "settings")
		return nil, false, nsutils.BadDataBodyError
	}

	if nsModel.Identifier.String() == "" {
		nsModel.Identifier = ptr(nightscout.IdentifierString(identifier))
	}

	if nsModel.Identifier.String() != identifier {
		slog.Error("setting identifier mismatch", "data", patch, "collection", "settings")
		return nil, false, nsutils.ErrBadRequest
	}

	sm, herr := UpdateIndividualSetting(ctx, siteContext, nsModel.Identifier.String(), nsModel.Value)
	if herr != nil {
		slog.Error("failed to update individual setting", "error", herr, "collection", "settings")
		return nil, false, herr
	}

	return ptr(c.converter(sm, siteContext, false)), false, nil
}

func (c SettingsCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	if identifier == "" {
		slog.Error("setting identifier not available", "data", patch, "collection", "settings")
		return nil, nsutils.ErrNotFound
	}
	if patch["identifier"] == nil {
		patch["identifier"] = identifier
	}

	record, _, err := c.Replace(ctx, siteContext, identifier, patch, false, raw)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (c SettingsCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	_, herr := UpdateIndividualSetting(ctx, siteContext, identifier, nil)
	if herr != nil {
		slog.Error("failed to update individual setting", "error", herr, "collection", "settings")
		return herr
	}

	return nil
}

func (c SettingsCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {

	setting, err := siteContext.GetSettings().GetSettingAnyTypeByName(identifier)
	if err != nil {
		return nil, nsutils.ErrNotFound
	}

	return ptr(c.converter(setting, siteContext, false)), nil
}

func (c SettingsCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {

	return nsutils.CreateRecordSlice(siteContext.GetSettings().GetAllSettingAnyTypes(), c.converter, siteContext, false), nil
}

func (c SettingsCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

///////
// SETTINGS TYPE
///////

type SettingAny struct {
	definition SettingDefinitionAny
	Value      any
	CreatedAt  int64
}

type SettingDefinitionAny interface {
	GetIdentifier() string
	Parse(any) (any, error)
	Validate(any) error
}

func (s *SettingAny) GetID() string {
	return s.definition.GetIdentifier()
}

func (s *SettingAny) GetDeletedAt() *int64 {
	return nil
}

func (s *SettingAny) GetCreatedAt() int64 {
	return s.CreatedAt
}

func (s *SettingAny) GetUpdatedAt() int64 {
	return s.CreatedAt
}

func (s *SettingAny) GetOccurredAt() int64 {
	return s.CreatedAt
}

func (s *SettingAny) GetCreatedByIdentityName() string {
	return ""
}

func (s *SettingAny) GetUtcOffsetMinutes() *int64 {
	return nil
}

func (s *SettingAny) GetCreatedByAppName() *string {
	return ptr("")
}

func (s *SettingAny) GetCreatedByDeviceName() *string {
	return ptr("")
}

func (s *SettingAny) GetRawJsonData() *string {
	return nil
}

func UpdateIndividualSetting(ctx context.Context, siteContext *httputils.SiteContext, identifier string, value any) (*settings.SettingAnyType, *nsutils.HTTPError) {
	settingsMap := siteContext.GetSettings()
	identitySettingsMap := settingsMap.IdentitySettingsMap

	err := identitySettingsMap.SetSettingByName(identifier, value, &siteContext.IdentityModel.CreatedAt, ptr("identity"))
	if err != nil {
		slog.Error("failed to set setting in identity settings map", "error", err, "collection", "settings")
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
		slog.Error("failed to update identity settings", "error", err, "collection", "settings")
		return nil, nsutils.ErrInternal
	}

	return identitySettingsMap.GetSettingAnyTypeByName(identifier), nil
}
