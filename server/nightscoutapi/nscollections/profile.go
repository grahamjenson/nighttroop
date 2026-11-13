package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"slices"
	"strings"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/types"

	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileCollection struct {
	named[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle]
}

func NewProfileCollection() CollectionStore {
	return NewCollection[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle](ProfileCollection{
		named[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle]{nsutils.Profile, DBToNightscoutProfileStore},
	})
}

func (c ProfileCollection) Validate(nsModel nightscoutmodels.ProfileStore, dbModel *sitedb.ProfilePresetBundle) *nsutils.HTTPError {
	// Every profile in the store becomes a profile_presets row, so cap the store before any are inserted
	if len(nsModel.Store) > consts.MaxProfilesPerStore {
		return nsutils.ErrTooManyProfiles
	}
	if dbModel == nil {
		return nil
	}

	return nil
}

// Rows are a preset for each of its profiles, by their names, as the store keeps them in no order
func (c ProfileCollection) Rows(nsModel nightscoutmodels.ProfileStore) []sitedb.TableRow {
	presets := NightscoutToDBCreateProfilePreset(nsModel, nightscoutmodels.ConverterContext{})
	slices.SortFunc(presets, func(a, b sitedb.CreateProfilePresetParams) int { return strings.Compare(a.Name, b.Name) })
	rows := make([]sitedb.TableRow, len(presets))
	for i, preset := range presets {
		rows[i] = sitedb.TableRow{Table: "profile_presets", Row: preset}
	}
	return rows
}

func (c ProfileCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.ProfileStore, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.ProfilePresetBundle, error) {
	// Loop's and Trio's settings are kept once in loop_settings, and each profile refers to them. Loop's glucose is
	// kept in mg/dL, as it sends it in its profile's units; Trio's is in mg/dL.
	sent := nightscoutmodels.PartialProfileStoreToCreate(nsModel, cc)
	loopSettings := nsModel.LoopSettings.WithGlucose(nsModel.Units.ToMGDL)
	loopID, err := recordLoopSettings(q, ctx, loopApp, fromProfile, loopSettings, sent.OccurredAt, sent.CreatedByDeviceName)
	if err != nil {
		return nil, err
	}
	remoteID, err := recordLoopSettings(q, ctx, trioApp, fromProfile, nsModel.RemoteSettings, sent.OccurredAt, sent.CreatedByDeviceName)
	if err != nil {
		return nil, err
	}

	ret := sitedb.ProfilePresetBundle([]*sitedb.ProfilePreset{})
	for _, createParams := range NightscoutToDBCreateProfilePreset(nsModel, cc) {
		// A replace keeps the store's identifier, as every preset in it shares it
		if overrideIdentifier != nil {
			createParams.ProfileStoreID = *overrideIdentifier
		}
		createParams.LoopSettingsID, createParams.RemoteSettingsID = loopID, remoteID
		dbModel, err := q.CreateProfilePreset(ctx, createParams)
		if err != nil {
			return nil, err
		}
		ret = append(ret, dbModel)
	}
	return &ret, nil
}

// FallbackKey is a profile store's date and device
func (c ProfileCollection) FallbackKey() FallbackKey[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle] {
	return FallbackKey[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle]{
		search: c.searchSameTimeStore,
		fields: []KeyField[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle]{
			sameDate[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle](),
			sameDevice[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle](),
		},
	}
}

// searchSameTimeStore finds the newest profile store from the device at the date of the one sent. It reads only that
// one, as each store is read with all of its presets.
func (c ProfileCollection) searchSameTimeStore(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.ProfileStore, _ int64) ([]*sitedb.ProfilePresetBundle, error) {
	profilePresets, err := q.SearchProfileStores(ctx, sitedb.SearchProfileStoresParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		Limit:    1,
	})
	if err != nil || len(profilePresets) == 0 {
		return nil, err
	}
	store, err := c.GetModel(q, ctx, profilePresets[0].ProfileStoreID, true)
	if err != nil {
		return nil, err
	}
	return []*sitedb.ProfilePresetBundle{store}, nil
}

func (c ProfileCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.ProfilePresetBundle, error) {
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

// Delete always deletes a profile store for good, as presets are not soft deleted
func (c ProfileCollection) Delete(q *sitedb.Queries, ctx context.Context, identifier string, _ bool) (sql.Result, error) {
	return q.HardDeleteProfileStore(ctx, sitedb.HardDeleteProfileStoreParams{ProfileStoreID: identifier})
}

func (c ProfileCollection) AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error) {
	return q.ProfileStoreAscendingIDs(ctx, nsquery.ProfileStoreAscending(query))
}

func (c ProfileCollection) HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error) {
	return q.ProfileStoreHistoryIDs(ctx, sitedb.ProfileStoreHistoryIDsParams{Since: since, Limit: limit})
}

func (c ProfileCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.ProfilePresetBundle, error) {
	var profilePresets []*sitedb.ProfilePreset
	var err error
	// With a count, as LoopFollow's remote diagnostics send, it is that many of the newest stores, as Nightscout gives
	if !query.HasFilters && !query.CountGiven {
		profilePresets, err = q.GetMostRecentProfileStore(ctx)
	} else {
		profilePresets, err = q.SearchProfileStores(ctx, nsquery.ProfileStoreSearch(query))
	}
	return sitedb.ManyPresetsToMayWrappers(profilePresets), err
}

///////
// DB TO NIGHTSCOUT
///////

func DBToNightscoutProfileStore(source *sitedb.ProfilePresetBundle, siteContext *httputils.SiteContext) nightscoutmodels.ProfileStore {
	if source == nil {
		return nightscoutmodels.ProfileStore{}
	}
	realSource := *source
	if len(realSource) == 0 {
		return nightscoutmodels.ProfileStore{}
	}
	defaultProfile := realSource[0]
	for _, p := range realSource {
		if p.IsDefault {
			defaultProfile = p
			break
		}
	}
	settingsUnits := siteContext.GetSettings().GetUnits()

	profileStore := nightscoutmodels.DBModelToProfileStore(defaultProfile, siteContext)
	profileStore.Store = make(map[types.ShortString]nightscoutmodels.Profile)
	profileStore.DefaultProfile = ptr(types.ShortString(defaultProfile.Name))
	profileStore.Units = settingsUnits
	// Loop's and Trio's settings are the loop_settings it refers to, left out when they were deleted. Loop's glucose
	// is in the profile's units, from mg/dL, or guessed from its size for settings kept in the units Loop sent.
	if loop := loopSettingsJSON(context.Background(), siteContext, defaultProfile.LoopSettingsID); loop != nil {
		profileStore.LoopSettings = nightscoutmodels.JSONToLoopSettings(ptr(string(loop))).WithGlucose(func(glucose float64) float64 {
			return settingsUnits.FromMGDL((*consts.Units)(nil).ToMGDL(glucose))
		})
	}
	if remote := loopSettingsJSON(context.Background(), siteContext, defaultProfile.RemoteSettingsID); remote != nil {
		profileStore.RemoteSettings = nightscoutmodels.JSONToRemoteSettings(ptr(string(remote)))
	}

	for _, p := range realSource {
		profile := nightscoutmodels.DBModelToProfile(p, siteContext)
		profile.Units = settingsUnits

		profileStore.Store[types.ShortString(p.Name)] = profile
	}

	return profileStore
}

///////
// NIGHTSCOUT TO DB
///////

func NightscoutToDBCreateProfilePreset(source nightscoutmodels.ProfileStore, ctx nightscoutmodels.ConverterContext) []sitedb.CreateProfilePresetParams {
	createParams := []sitedb.CreateProfilePresetParams{}
	profileStoreID := nightscoutmodels.AssignID()
	defaultUnits := source.Units
	for name, profile := range source.Store {
		createParam := nightscoutmodels.PartialProfileStoreToCreate(source, ctx)
		createParam.ProfileStoreID = profileStoreID
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
