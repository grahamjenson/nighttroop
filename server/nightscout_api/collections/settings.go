package collections

import (
	"context"
	"reflect"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type SettingsCollection struct {
	converter func(c *Setting) nsutils.Record
}

func NewSettingsCollection() SettingsCollection {
	return SettingsCollection{
		converter: nsutils.DBModelToRecord(nsutils.Settings, ConvertFromNSToSetting),
	}
}

func (c SettingsCollection) Name() string {
	return "settings"
}

func (c SettingsCollection) Convert(dr *Setting) nsutils.Record {
	return c.converter(dr)
}

// Cannot change properties

func (c SettingsCollection) Validate(nsModel nightscout.Setting, dbModel *Setting) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	return nil

}

func (c SettingsCollection) ValidatePatch(nsModel nightscout.Setting, dbModel *Setting) *nsutils.HTTPError {
	return c.Validate(nsModel, dbModel)
}

func (c SettingsCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Setting, cc nightscout.ConverterContext, overrideIdentifier *string) (*Setting, error) {
	siteContext := httputils.GetSiteContext(ctx)
	// if creating default need settings:admin permissions
	setting := Setting{
		Identifier: "asd",
		identity:   siteContext.IdentityModel,
		site:       siteContext.SiteModel,
	}
	return &setting, nil

}

func (c SettingsCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*Setting, error) {
	siteContext := httputils.GetSiteContext(ctx)
	settingsMap := GetSettingsMap(siteContext.SiteModel, siteContext.IdentityModel)
	setting, ok := settingsMap[identifier]
	if !ok {
		return nil, nsutils.ErrNotFound
	}
	return &setting, nil
}

func (c SettingsCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Setting, identifier string, dbModel *Setting, uc nightscout.ConverterContext) (*Setting, error) {
	siteContext := httputils.GetSiteContext(ctx)
	settingsMap := GetSettingsMap(siteContext.SiteModel, siteContext.IdentityModel)
	setting, ok := settingsMap[identifier]
	if !ok {
		return nil, nsutils.ErrNotFound
	}
	return &setting, nil
}

// ///////
// HTTP funcs
// ///////
func (c SettingsCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c SettingsCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
}

func (c SettingsCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	settingsMap := GetSettingsMap(siteContext.SiteModel, siteContext.IdentityModel)
	settingsList := make([]*Setting, 0, len(settingsMap))
	for _, setting := range settingsMap {
		settingsList = append(settingsList, &setting)
	}
	return nsutils.CreateRecordSlice(settingsList, c.converter), nil
}

func (c SettingsCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c SettingsCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c SettingsCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	record, err := c.Patch(ctx, siteContext, identifier, patch)
	if err != nil {
		return nil, false, err
	}
	return record, false, nil
}

func (c SettingsCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch)
}

///////
// SETTINGS TYPE
///////

func ConvertFromNSToSetting(s *Setting) nightscout.Setting {
	return nightscout.Setting{
		Identifier: ptr(nightscout.IdentifierString(s.Identifier)),
		Value:      s.Value,
		Type:       ptr(nightscout.ShortString(s.Type)),
	}
}

type Setting struct {
	Identifier string `json:"identifier"`
	Value      any    `json:"value"`
	Type       string `json:"type"`

	isDefault bool              `json:"-"`
	site      *control.Site     `json:"-"`
	identity  *control.Identity `json:"-"`
}

func (s Setting) GetID() string {
	return s.Identifier
}

func (s Setting) GetDeletedAt() *int64 {
	return nil
}

func (s Setting) GetCreatedAt() int64 {
	if s.isDefault {
		return s.site.CreatedAt
	}
	return s.identity.CreatedAt
}

func (s Setting) GetUpdatedAt() int64 {
	if s.isDefault {
		return s.site.CreatedAt
	}
	return s.identity.CreatedAt
}

func (s Setting) GetCreatedByIdentityName() string {
	return ""
}

func (s Setting) GetUtcOffsetMinutes() *int64 {
	return nil
}

func (s Setting) GetOccurredAt() int64 {
	if s.isDefault {
		return s.site.CreatedAt
	}
	return s.identity.CreatedAt
}

func (s Setting) GetCreatedByAppName() *string {
	return ptr("")
}

func (s Setting) GetCreatedByDeviceName() *string {
	return ptr("")
}

func GetSettingsMap(site *control.Site, identity *control.Identity) map[string]Setting {
	settings := make(map[string]Setting)
	for _, setting := range control.ALL_SETTINGS {
		defaultValue := control.ALL_SETTINGS_DEFAULTS[setting]
		settings[setting] = NewSetting(
			setting,
			identity.GetSetting(setting),
			site.GetSetting(setting),
			defaultValue,
			control.ALL_SETTINGS_TYPES[setting],
			site,
			identity,
		)
	}

	return settings
}

func NewSetting(
	identifier string,
	identityValue any,
	siteValue any,
	defaultValue any,
	typ string,
	site *control.Site,
	identity *control.Identity,
) Setting {
	return Setting{
		Identifier: identifier,
		Value:      first(identityValue, siteValue, defaultValue),
		Type:       typ,
		isDefault:  identityValue == nil,
		site:       site,
		identity:   identity,
	}
}

func first(values ...any) any {
	for _, value := range values {
		if value == nil {
			continue
		}

		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
			if v.IsNil() {
				continue
			}
		}

		return value
	}

	return nil
}
