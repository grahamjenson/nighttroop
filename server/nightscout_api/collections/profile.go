package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileCollection struct {
	converter func(c *sitedb.ProfilePresetBundle, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewProfileCollection() ProfileCollection {
	return ProfileCollection{converter: nsutils.DBModelToRecord(nsutils.Profile, DBToNightscoutProfileStore)}
}

func (c ProfileCollection) Name() string {
	return "profile"
}

func (c ProfileCollection) Convert(dr *sitedb.ProfilePresetBundle, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

func (c ProfileCollection) Validate(nsModel nightscout.ProfileStore, dbModel *sitedb.ProfilePresetBundle) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	return nil
}

func (c ProfileCollection) CreateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscout.ProfileStore, cc nightscout.ConverterContext, overrideIdentifier *string) (*sitedb.ProfilePresetBundle, error) {
	manyCreateParams := NightscoutToDBCreateProfilePreset(nsModel, cc)

	ret := sitedb.ProfilePresetBundle([]*sitedb.ProfilePreset{})

	for _, createParams := range manyCreateParams {
		dbModel, err := q.CreateProfilePreset(ctx, createParams)
		if err != nil {
			return nil, err
		}
		ret = append(ret, dbModel)
	}
	return &ret, nil
}

func (c ProfileCollection) GetDuplicateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscout.ProfileStore) (*sitedb.ProfilePresetBundle, error) {
	profilePresets, rerr := q.SearchProfileStores(ctx, sitedb.SearchProfileStoresParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(profilePresets) > 0 {
		identifier := profilePresets[0].ProfileStoreID
		return c.GetDBFunc(q, ctx, identifier, true)
	}
	return nil, nil
}

func (c ProfileCollection) GetDBFunc(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.ProfilePresetBundle, error) {
	ret, err := q.GetProfileStore(ctx, sitedb.GetProfileStoreParams{ProfileStoreID: identifier})
	if err != nil {
		return nil, err
	}
	wrapper := sitedb.ProfilePresetBundle(ret)
	if len(wrapper) == 0 {
		return nil, sql.ErrNoRows
	}
	return &wrapper, nil
}

func (c ProfileCollection) HardDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteProfileStore(ctx, sitedb.HardDeleteProfileStoreParams{ProfileStoreID: identifier})
}

func (c ProfileCollection) SoftDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteProfileStore(ctx, sitedb.HardDeleteProfileStoreParams{ProfileStoreID: identifier})
}

// ////
// HTTP METHODS
// ////
func (c ProfileCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c ProfileCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw, query.IncludeDeleted)
}

func (c ProfileCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var profilePresets []*sitedb.ProfilePreset
	var rerr error

	// There are three options,
	if !query.HasFilters {
		profilePresets, rerr = siteContext.SiteDB.Queries.GetMostRecentProfileStore(ctx)
	} else {
		profilePresets, rerr = siteContext.SiteDB.Queries.SearchProfileStores(ctx, nsquery.ProfileStoreSearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query profile presets from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}
	wrappers := sitedb.ManyPresetsToMayWrappers(profilePresets)
	return nsutils.CreateRecordSlice[*sitedb.ProfilePresetBundle](wrappers, c.converter, siteContext, query.Raw), nil
}

func (c ProfileCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c ProfileCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate, raw)
}

func (c ProfileCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

///////
// DB TO NIGHTSCOUT
///////

func DBToNightscoutProfileStore(source *sitedb.ProfilePresetBundle, siteContext *httputils.SiteContext) nightscout.ProfileStore {
	if source == nil {
		return nightscout.ProfileStore{}
	}
	realSource := *source
	if len(realSource) == 0 {
		return nightscout.ProfileStore{}
	}
	defaultProfile := realSource[0]
	for _, p := range realSource {
		if p.IsDefault {
			defaultProfile = p
			break
		}
	}
	settingsUnits := siteContext.GetSettings().GetUnits()

	profileStore := nightscout.DBModelToProfileStore(defaultProfile, siteContext)
	profileStore.Store = make(map[nightscout.ShortString]nightscout.Profile)
	profileStore.DefaultProfile = ptr(nightscout.ShortString(defaultProfile.Name))
	profileStore.Units = settingsUnits

	for _, p := range realSource {
		profile := nightscout.DBModelToProfile(p, siteContext)
		profile.Units = settingsUnits

		profileStore.Store[nightscout.ShortString(p.Name)] = profile
	}

	return profileStore
}

///////
// NIGHTSCOUT TO DB
///////

func NightscoutToDBCreateProfilePreset(source nightscout.ProfileStore, ctx nightscout.ConverterContext) []sitedb.CreateProfilePresetParams {
	createParams := []sitedb.CreateProfilePresetParams{}
	profileStoreID := nightscout.AssignID()
	var loopSettingsJSON *string
	if source.LoopSettings != nil {
		loopSettingsJSON = TryValueToJSON(source.LoopSettings)
	}
	defaultUnits := source.Units
	for name, profile := range source.Store {
		createParam := nightscout.PartialProfileStoreToCreate(source, ctx)
		createParam.ProfileStoreID = profileStoreID
		createParam.LoopSettingsJson = loopSettingsJSON
		createParam.Name = string(name)
		createParam.IsDefault = source.DefaultProfile != nil && *source.DefaultProfile == name

		createParam.Dia = profile.Dia.Float64p()
		createParam.Timezone = profile.Timezone.Stringp()
		createParam.CarbsHr = profile.CarbsHr.Float64p()
		createParam.Delay = profile.Delay.Float64p()

		if profile.Units == nil {
			profile.Units = defaultUnits
		}

		createParam.Units = profile.Units

		profile.ConvertToMGDL() // Converts the target values to

		if len(profile.CarbRatio) > 0 {
			createParam.CarbRatiosJson = TryValueToJSON(profile.CarbRatio)
		}
		if len(profile.Sens) > 0 {
			createParam.InsulinSensitivitiesJson = TryValueToJSON(profile.Sens)
		}
		if len(profile.Basal) > 0 {
			createParam.BasalRatesJson = TryValueToJSON(profile.Basal)
		}
		if len(profile.TargetLow) > 0 {
			createParam.TargetsLowJson = TryValueToJSON(profile.TargetLow)
		}
		if len(profile.TargetHigh) > 0 {
			createParam.TargetsHighJson = TryValueToJSON(profile.TargetHigh)
		}

		createParams = append(createParams, createParam)
	}

	return createParams
}

func TryValueToJSON(svs any) *string {

	jsonBytes, err := json.Marshal(svs)
	if err != nil {
		slog.Error("error marshalling profile values to JSON", "error", err)
		return nil
	}
	jsonStr := string(jsonBytes)
	return &jsonStr
}
