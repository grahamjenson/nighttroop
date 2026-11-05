package sitesettings

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
)

// SettingsCollection is a level of settings: the site's, which only admins can change,
// or the signed in identity's own, which override the site's. Alarms shows the identity's on the alarms page.
// Other is another identity's own, by its ID, which an admin changes for it, such as a device's units.
type SettingsCollection struct {
	Site   bool
	Alarms bool
	Other  bool
}

// settingNames are the settings the pages show so far, in order, and the units labels, which only the endpoint sets
var settingNames = []string{"units", "units_mgdl_label", "units_mmol_label", "timezone", "language", "date_format", "time_format", "week_start_day",
	"alarm_urgent_low_enabled", "range_urgent_low", "alarm_low_enabled", "range_low", "alarm_low_minutes",
	"alarm_low_prediction_enabled", "alarm_low_prediction_minutes",
	"alarm_high_enabled", "range_high", "alarm_high_minutes", "alarm_urgent_high_enabled", "range_urgent_high",
	"alarm_fast_rise_enabled", "alarm_fast_rise", "alarm_fast_rise_minutes", "alarm_fast_drop_enabled", "alarm_fast_drop",
	"alarm_fast_drop_minutes", "alarm_no_data_enabled", "alarm_no_data_minutes"}

// glucoseSettings are shown and typed in the settings' units, a fast rise or drop as how far, and kept in mg/dL as
// the setting of the same name ending _mgdl
var glucoseSettings = []string{"range_urgent_low", "range_low", "range_high", "range_urgent_high", "alarm_fast_rise", "alarm_fast_drop"}

// settingID is the setting a field on the page is kept as
func settingID(name string) string {
	if slices.Contains(glucoseSettings, name) {
		return name + "_mgdl"
	}
	return name
}

// showSetting is a setting's value as the page shows it, one kept in mg/dL in the settings' units
func showSetting(setting *settings.SettingAnyType, units *consts.Units) string {
	if mgdl, ok := setting.Value.(int64); ok && strings.HasSuffix(setting.Identifier, "_mgdl") {
		return site_data_templ.FormatGlucose(units.FromMGDL(float64(mgdl)), units)
	}
	return fmt.Sprint(setting.Value)
}

// ////
// Shared
// ////
func (s *SettingsCollection) GetName() string {
	if s.Other {
		return "other_identity_settings"
	}
	if s.Site {
		return "site_settings"
	}
	if s.Alarms {
		return "alarm_settings"
	}
	return "identity_settings"
}

// GetRoute is where the collection is in its URL
func (s *SettingsCollection) GetRoute() string {
	if s.Other {
		return "identity-settings"
	}
	if s.Site {
		return "site"
	}
	if s.Alarms {
		return "alarms"
	}
	return "identity"
}

func (s *SettingsCollection) level(siteContext *httputils.SiteContext, identity *controldb.SiteIdentity) *settings.SettingMap {
	if s.Site {
		return siteContext.GetSettings().SiteSettingsMap
	}
	if s.Other {
		return settings.NewSettingsMap(identity, &identity.CreatedAt, new("identity"))
	}
	return siteContext.GetSettings().IdentitySettingsMap
}

// identity is whose settings these are: the one with the ID, of the signed in site, for Other, or the one signed in
func (s *SettingsCollection) identity(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*controldb.SiteIdentity, error) {
	if !s.Other {
		return siteContext.IdentityModel, nil
	}
	identityID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, apiutils.ErrNotFound
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	return q.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.SiteID, ID: identityID})
}

// ////
// Get Model
// /////
// GetRow is the level's settings, with those it inherits; settings are not rows, so the model is its own row
func (s *SettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Settings, error) {
	if s.Site && siteContext.IdentityModel.Role != consts.IdentityRoleAdmin {
		return nil, apiutils.ErrForbidden
	}
	identity, err := s.identity(ctx, collectionTx, id, siteContext)
	if err != nil {
		return nil, err
	}
	// A glucose is shown in the units of the identity signed in, whoever's settings these are, as its label says
	combined, units, level := siteContext.GetSettings(), siteContext.GetSettings().GetUnits(), s.level(siteContext, identity)
	model := &sitemodels.Settings{Inherited: map[string]string{}}
	if !s.Site {
		model.OwnerID = &identity.ID
	}
	if s.Other {
		model.IdentityID, model.IdentityName = &identity.ID, &identity.Name
	}
	fields := model.Fields()
	for _, name := range settingNames {
		if setting := level.GetSettingAnyTypeByName(settingID(name)); setting != nil {
			*fields[name] = new(showSetting(setting, units))
		}
		inherited := combined.DefaultSettingsMap.GetSettingAnyTypeByName(settingID(name))
		if site := combined.SiteSettingsMap.GetSettingAnyTypeByName(settingID(name)); !s.Site && site != nil {
			inherited = site
		}
		model.Inherited[name] = showSetting(inherited, units)
	}

	// An admin's settings page shows the site's settings beside their own
	if !s.Site && !s.Other && siteContext.IdentityModel.Role == consts.IdentityRoleAdmin {
		site, err := (&SettingsCollection{Site: true}).GetRow(ctx, collectionTx, id, siteContext)
		if err != nil {
			return nil, err
		}
		model.Site = site
	}
	return model, nil
}

// Watches is whether a settings change is streamed to the identity: every level is the same model, so its own
// settings and alarms are only its own changes and, for an admin, the site's, beside which an admin's are shown.
// The site's and other identities' are the admins' alone, as reading them is.
func (s *SettingsCollection) Watches(event events.Event, siteContext *httputils.SiteContext) bool {
	changed, ok := event.Model.(*sitemodels.Settings)
	if !ok {
		return false
	}
	switch {
	case s.Site:
		return changed.OwnerID == nil
	case s.Other:
		return changed.OwnerID != nil
	case changed.OwnerID == nil:
		return siteContext.IdentityModel.Role == consts.IdentityRoleAdmin
	}
	return *changed.OwnerID == siteContext.IdentityModel.ID
}

func (s *SettingsCollection) ToModel(row *sitemodels.Settings, _ *httputils.SiteContext) *sitemodels.Settings {
	return row
}

func (s *SettingsCollection) GetHTMXResponse(model *sitemodels.Settings) templ.Component {
	if s.Other {
		return site_settings_templ.IdentitySettingsPage(model)
	}
	page := site_settings_templ.SettingsPage
	if s.Alarms {
		page = site_settings_templ.AlarmSettingsPage
	}
	if s.Site {
		return page(model, nil)
	}
	return page(model.Site, model)
}

// ///
// Update
// ///

// UpdateRow sets the values sent. An empty value clears it, so the level above applies; one not sent is left as it is.
func (s *SettingsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.Settings, model *sitemodels.Settings) (*sitemodels.Settings, error) {
	// Another identity's are those of the one read, not of any ID sent
	id := ""
	if oldModel.IdentityID != nil {
		id = strconv.FormatInt(*oldModel.IdentityID, 10)
	}
	identity, err := s.identity(ctx, collectionTx, id, siteContext)
	if err != nil {
		return nil, err
	}
	level, subject := s.level(siteContext, identity), new("identity")
	if s.Site {
		subject = new("site")
	}
	sent, units := model.Fields(), siteContext.GetSettings().GetUnits()
	for _, name := range settingNames {
		if *sent[name] == nil {
			continue
		}
		var value any
		if typed := **sent[name]; typed != "" {
			value = typed
		}
		// A glucose is typed in the units it was shown in, and kept as whole mg/dL
		if typed, ok := value.(string); ok && slices.Contains(glucoseSettings, name) {
			glucose, err := strconv.ParseFloat(typed, 64)
			if err != nil {
				return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("%s must be a number", name))
			}
			value = int64(math.Round(units.ToMGDL(glucose)))
		}
		if err := level.SetSettingByName(settingID(name), value, nil, subject); err != nil {
			return nil, apiutils.ErrBadBodyWithHint(err.Error())
		}
	}

	// The settings in effect must make sense together: the identity's over the site's, or the site's over the
	// defaults for everyone else
	inEffect := siteContext.GetSettings()
	if s.Site {
		inEffect = &settings.CombinedSettingsMap{SiteSettingsMap: level, IdentitySettingsMap: &settings.SettingMap{}, DefaultSettingsMap: settings.DefaultSettingsMap}
	}
	if s.Other {
		inEffect = &settings.CombinedSettingsMap{SiteSettingsMap: inEffect.SiteSettingsMap, IdentitySettingsMap: level, DefaultSettingsMap: settings.DefaultSettingsMap}
	}
	if err := inEffect.Consistent(); err != nil {
		return nil, apiutils.ErrBadBodyWithHint(err.Error())
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	// Every setting is written, so the ones not on the page keep their values
	if s.Site {
		params := &controldb.UpdateSitesSettingsParams{ID: siteContext.SiteModel.ID}
		settings.SetSettings(params, level)
		err = q.UpdateSitesSettings(ctx, *params)
	} else {
		params := &controldb.UpdateSiteIdentitySettingsParams{ID: identity.ID, SiteID: siteContext.SiteID}
		settings.SetSettings(params, level)
		err = q.UpdateSiteIdentitySettings(ctx, *params)
	}
	if err != nil {
		return nil, err
	}

	// A new site timezone is an event, once saved, so the background jobs that follow it, such as the site's days,
	// see it
	if was, now := siteContext.GetSiteTimezone().String(), inEffect.Location().String(); s.Site && now != was {
		collectionTx.OnCommit(func() {
			siteContext.ControlDB.Logs.SiteTimezoneChangedEvent(ctx, siteContext.SiteID, siteContext.IdentityModel.Name, was, now)
		})
	}
	if was, now := siteContext.GetSiteWeekStartDay(), *inEffect.GetWeekStartDay(); s.Site && now != was {
		collectionTx.OnCommit(func() {
			siteContext.ControlDB.Logs.SiteWeekStartChangedEvent(ctx, siteContext.SiteID, siteContext.IdentityModel.Name, was, now)
		})
	}
	return s.GetRow(ctx, collectionTx, id, siteContext)
}

func (s *SettingsCollection) UpdateHTMXResponse(model *sitemodels.Settings) templ.Component {
	return site_settings_templ.SettingsSaved()
}
