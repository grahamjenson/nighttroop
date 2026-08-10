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
	converter func(c *ProfilePresetBundle) nsutils.Record
}

func NewProfileCollection() ProfileCollection {
	return ProfileCollection{converter: nsutils.DBModelToRecord(nsutils.Profile, DBToNightscout)}
}

func (c ProfileCollection) Name() string {
	return "profile"
}

func (c ProfileCollection) Convert(dr *ProfilePresetBundle) nsutils.Record {
	return c.converter(dr)
}

func (c ProfileCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.ProfileStore, cc nightscout.CreateContext, replacing *ProfilePresetBundle, overrideIdentifier *string) (*ProfilePresetBundle, error) {
	manyCreateParams := NightscoutToDBCreate(nsModel, cc)
	// Validate change if replacing a model
	if replacing != nil {
		// TODO: Validate something here
	}

	ret := ProfilePresetBundle([]*site.ProfilePreset{})

	for _, createParams := range manyCreateParams {
		dbModel, err := q.CreateProfilePreset(ctx, createParams)
		if err != nil {
			return nil, err
		}
		ret = append(ret, dbModel)
	}
	return &ret, nil
}

func (c ProfileCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*ProfilePresetBundle, error) {
	ret, err := q.GetProfileStore(ctx, site.GetProfileStoreParams{ProfileStoreID: identifier})
	if err != nil {
		return nil, err
	}
	wrapper := ProfilePresetBundle(ret)
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
func (c ProfileCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c ProfileCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
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
	wrappers := ManyPresetsToMayWrappers(profilePresets)
	return nsutils.CreateRecordSlice[*ProfilePresetBundle](wrappers, c.converter), nil
}

func (c ProfileCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c ProfileCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate)
}

func (c ProfileCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

///////////////////////////
// CONVERSION METHOD
///////////////////////////

type ProfilePresetBundle []*site.ProfilePreset

func (c ProfilePresetBundle) GetID() string {
	if len(c) == 0 || c[0] == nil {
		return ""
	}
	return c[0].ProfileStoreID
}
func (c ProfilePresetBundle) GetDeletedAt() *int64 {
	return nil

}
func (c ProfilePresetBundle) GetCreatedAt() int64 {
	if len(c) == 0 || c[0] == nil {
		return 0
	}
	return c[0].CreatedAt
}
func (c ProfilePresetBundle) GetUpdatedAt() int64 {
	if len(c) == 0 || c[0] == nil {
		return 0
	}
	return c[0].UpdatedAt
}

func (c ProfilePresetBundle) GetCreatedByIdentityName() string {
	if len(c) == 0 || c[0] == nil {
		return ""
	}
	return c[0].CreatedByIdentityName
}

func (c ProfilePresetBundle) GetUtcOffsetMinutes() *int64 {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].UtcOffsetMinutes
}
func (c ProfilePresetBundle) GetOccurredAt() int64 {
	if len(c) == 0 || c[0] == nil {
		return 0
	}
	return c[0].OccurredAt
}

func (c ProfilePresetBundle) GetCreatedByAppName() *string {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].CreatedByAppName
}

func (c ProfilePresetBundle) GetCreatedByDeviceName() *string {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].CreatedByDeviceName
}

///////
// Many to Many
///////

func ManyPresetsToMayWrappers(source []*site.ProfilePreset) []*ProfilePresetBundle {
	if source == nil {
		return nil
	}
	lookup := map[string][]*site.ProfilePreset{}
	for _, preset := range source {
		if preset == nil {
			continue
		}
		if _, ok := lookup[preset.ProfileStoreID]; !ok {
			lookup[preset.ProfileStoreID] = []*site.ProfilePreset{}
		}
		if preset.IsDefault {
			// If this is the default profile, we want it to be first in the list
			lookup[preset.ProfileStoreID] = append([]*site.ProfilePreset{preset}, lookup[preset.ProfileStoreID]...)

		} else {
			lookup[preset.ProfileStoreID] = append(lookup[preset.ProfileStoreID], preset)
		}
	}

	wrappers := []*ProfilePresetBundle{}
	for _, presets := range lookup {
		wrapper := ProfilePresetBundle(presets)
		wrappers = append(wrappers, &wrapper)
	}
	return wrappers
}

///////
// DB TO NIGHTSCOUT
///////

func DBToNightscout(source *ProfilePresetBundle) nightscout.ProfileStore {
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

func NightscoutToDBCreate(source nightscout.ProfileStore, ctx nightscout.CreateContext) []site.CreateProfilePresetParams {
	createParams := []site.CreateProfilePresetParams{}
	profileStoreID := nightscout.AssignID()
	for name, profile := range source.Store {
		createParam := nightscout.PartialProfileStoreToCreate(source, ctx)
		createParam.ProfileStoreID = profileStoreID

		createParam.Name = string(name)
		createParam.IsDefault = source.DefaultProfile != nil && *source.DefaultProfile == name

		if profile.Dia != nil {
			createParam.Dia = ptr(profile.Dia.Float64())
		}

		createParam.Units = profile.Units.Stringp()
		createParam.Timezone = profile.Timezone.Stringp()

		if profile.CarbRatio != nil {
			createParam.CarbRatiosJson = ScheduleValuesToJson(profile.CarbRatio)
		}
		if profile.Sens != nil {
			createParam.InsulinSensitivitiesJson = ScheduleValuesToJson(profile.Sens)
		}
		if profile.Basal != nil {
			createParam.BasalRatesJson = ScheduleValuesToJson(profile.Basal)
		}
		if profile.TargetLow != nil {
			createParam.TargetsLowJson = ScheduleValuesToJson(profile.TargetLow)
		}
		if profile.TargetHigh != nil {
			createParam.TargetsHighJson = ScheduleValuesToJson(profile.TargetHigh)
		}

		createParams = append(createParams, createParam)
	}

	return createParams
}

///////
// ProfileStore To Profile Preset
///////

// func ProfileStoreToSingleCreate(source ProfileStore, name string, profile Profile, ctx CreateContext) site.CreateProfilePresetParams {
// 	params := site.CreateProfilePresetParams{
// 		Name:                  name,
// 		IsDefault:             source.DefaultProfile != nil && *source.DefaultProfile == ShortString(name),
// 		CreatedByAppName:      ctx.CreatedByAppName,
// 		CreatedByDeviceName:   ctx.CreatedByDeviceName,
// 		CreatedByIdentityName: ctx.CreatedByIdentityName,
// 	}

// 	if profile.Dia != nil {
// 		params.Dia = *profile.Dia
// 	}
// 	if profile.CarbRatio != nil {
// 		params.CarbRatiosJson = ScheduleValuesToJson(profile.CarbRatio)
// 	}
// 	if profile.Sens != nil {
// 		params.InsulinSensitivitiesJson = ScheduleValuesToJson(profile.Sens)
// 	}
// 	if profile.Basal != nil {
// 		params.BasalRatesJson = ScheduleValuesToJson(profile.Basal)
// 	}
// 	if profile.TargetLow != nil || profile.TargetHigh != nil {
// 		params.TargetsJson = TargetsToJson(profile.TargetLow, profile.TargetHigh)
// 	}
// 	if profile.Timezone != nil {
// 		params.Timezone = string(*profile.Timezone)
// 	}
// 	if profile.Units != nil {
// 		params.Units = string(*profile.Units)
// 	}

// 	return params
// }

// func ProfileStoreToCreateAll(source ProfileStore, profile Profile, ctx CreateContext) []site.CreateProfilePresetParams {
// 	creates := []site.CreateProfilePresetParams{}
// 	for profileName, profile := range source.Store {
// 		create := ProfileStoreToSingleCreate(source, string(profileName), profile, ctx)
// 		creates = append(creates, create)
// 	}
// 	return creates
// }

func ScheduleValuesToJson(svs []*nightscout.ScheduleValue) *string {
	if len(svs) == 0 {
		return nil
	}
	jsonBytes, err := json.Marshal(svs)
	if err != nil {
		slog.Error("error marshalling schedule values to JSON", "error", err)
		return nil
	}
	jsonStr := string(jsonBytes)
	return &jsonStr
}
