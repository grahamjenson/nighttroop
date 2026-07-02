package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type SettingsCollection struct {
	converter func(c *user.Setting) nsutils.Record
}

func NewSettingsCollection() SettingsCollection {
	return SettingsCollection{
		converter: DBModelToRecord(nsutils.Settings, nightscout.DBModelToSetting),
	}
}

func (c SettingsCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	settingModel := &nightscout.Setting{}
	err := utils.FromMap(data, settingModel)

	if err != nil {
		slog.Error("failed to map data to setting model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	key := settingModel.Identifier
	stringValue := settingModel.Value
	settingValue, ok := user.GetUserSettingFromKey(key.String())
	if !ok {
		slog.Error("invalid setting key", "key", key)
		return nil, nsutils.BadDataBodyError
	}

	value, err := settingValue.CreateValue(stringValue.String())
	if err != nil {
		slog.Error("failed to create setting value from string", "error", err)
		return nil, nsutils.BadDataBodyError
	}
	dbModel, err := settingValue.UpsertValue(ctx, userContext.UserDB.Queries, value)

	if err != nil {
		slog.Error("failed to create setting in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create setting in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converter(dbModel)), nil
}

func (c SettingsCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	setting, ok := user.GetUserSettingFromKey(identifier)
	if !ok {
		return nil, nsutils.ErrNotFound
	}
	dbModel, err := setting.GetValueModel(ctx, userContext.UserDB.Queries)
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get setting", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	record := c.converter(dbModel)
	return &record, nil
}

func (c SettingsCollection) Search(ctx context.Context, userContext *httputils.UserContext, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c SettingsCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c SettingsCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c SettingsCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
