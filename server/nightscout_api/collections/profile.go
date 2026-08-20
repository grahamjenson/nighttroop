package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileCollection struct {
	converter func(c *site.ProfilePresetBundle, raw bool) nsutils.Record
}

func NewProfileCollection() ProfileCollection {
	return ProfileCollection{converter: nsutils.DBModelToRecord(nsutils.Profile, DBToNightscoutProfileStore)}
}

func (c ProfileCollection) Name() string {
	return "profile"
}

func (c ProfileCollection) Convert(dr *site.ProfilePresetBundle, raw bool) nsutils.Record {
	return c.converter(dr, raw)
}

func (c ProfileCollection) Validate(nsModel nightscout.ProfileStore, dbModel *site.ProfilePresetBundle) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	return nil
}

func (c ProfileCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.ProfileStore, cc nightscout.ConverterContext, overrideIdentifier *string) (*site.ProfilePresetBundle, error) {
	manyCreateParams := NightscoutToDBCreateProfilePreset(nsModel, cc)

	ret := site.ProfilePresetBundle([]*site.ProfilePreset{})

	for _, createParams := range manyCreateParams {
		dbModel, err := q.CreateProfilePreset(ctx, createParams)
		if err != nil {
			return nil, err
		}
		ret = append(ret, dbModel)
	}
	return &ret, nil
}

func (c ProfileCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.ProfilePresetBundle, error) {
	ret, err := q.GetProfileStore(ctx, site.GetProfileStoreParams{ProfileStoreID: identifier})
	if err != nil {
		return nil, err
	}
	wrapper := site.ProfilePresetBundle(ret)
	return &wrapper, nil
}

func (c ProfileCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteProfileStore(ctx, site.HardDeleteProfileStoreParams{ProfileStoreID: identifier})
}

func (c ProfileCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteProfileStore(ctx, site.HardDeleteProfileStoreParams{ProfileStoreID: identifier})
}

// ////
// HTTP METHODS
// ////
func (c ProfileCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c ProfileCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw)
}

func (c ProfileCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var profilePresets []*site.ProfilePreset
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
	wrappers := site.ManyPresetsToMayWrappers(profilePresets)
	return nsutils.CreateRecordSlice[*site.ProfilePresetBundle](wrappers, c.converter, query.Raw), nil
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

func DBToNightscoutProfileStore(source *site.ProfilePresetBundle) nightscout.ProfileStore {
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

	profileStore := nightscout.DBModelToProfileStore(defaultProfile)
	profileStore.Store = make(map[nightscout.ShortString]nightscout.Profile)
	profileStore.DefaultProfile = ptr(nightscout.ShortString(defaultProfile.Name))
	if defaultProfile.Units != nil {
		profileStore.Units = ptr(nightscout.Units(*defaultProfile.Units))
	}
	for _, p := range realSource {
		profile := nightscout.DBModelToProfile(p)
		profileStore.Store[nightscout.ShortString(p.Name)] = profile
	}

	return profileStore
}

///////
// NIGHTSCOUT TO DB
///////

func NightscoutToDBCreateProfilePreset(source nightscout.ProfileStore, ctx nightscout.ConverterContext) []site.CreateProfilePresetParams {
	createParams := []site.CreateProfilePresetParams{}
	profileStoreID := nightscout.AssignID()
	var loopSettingsJSON *string
	if source.LoopSettings != nil {
		loopSettingsJSON = TryValueToJSON(source.LoopSettings)
	}

	for name, profile := range source.Store {
		createParam := nightscout.PartialProfileStoreToCreate(source, ctx)
		createParam.ProfileStoreID = profileStoreID
		createParam.LoopSettingsJson = loopSettingsJSON
		createParam.Name = string(name)
		createParam.IsDefault = source.DefaultProfile != nil && *source.DefaultProfile == name

		if profile.Dia != nil {
			createParam.Dia = ptr(profile.Dia.Float64())
		}

		createParam.Units = profile.Units.Stringp()
		createParam.Timezone = profile.Timezone.Stringp()

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
