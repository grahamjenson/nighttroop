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
// or the signed in identity's own, which override the site's
type SettingsCollection struct {
	Site bool
}

// settingNames are the settings the pages show so far, in order
var settingNames = []string{"units", "timezone", "language", "date_format", "time_format", "week_start_day",
	"alarm_urgent_low_enabled", "alarm_urgent_low", "alarm_low_enabled", "alarm_low", "alarm_low_minutes",
	"alarm_low_prediction_enabled", "alarm_low_prediction_minutes",
	"alarm_high_enabled", "alarm_high", "alarm_high_minutes", "alarm_urgent_high_enabled", "alarm_urgent_high",
	"alarm_fast_rise_enabled", "alarm_fast_rise", "alarm_fast_rise_minutes", "alarm_fast_drop_enabled", "alarm_fast_drop",
	"alarm_fast_drop_minutes", "alarm_no_data_enabled", "alarm_no_data_minutes"}

// glucoseSettings are shown and typed in the settings' units, a fast rise or drop as how far, and kept in mg/dL as
// the setting of the same name ending _mgdl
var glucoseSettings = []string{"alarm_urgent_low", "alarm_low", "alarm_high", "alarm_urgent_high", "alarm_fast_rise", "alarm_fast_drop"}

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
	if s.Site {
		return "site_settings"
	}
	return "identity_settings"
}

// GetRoute is where the collection is in its URL
func (s *SettingsCollection) GetRoute() string {
	if s.Site {
		return "site"
	}
	return "identity"
}

func (s *SettingsCollection) level(siteContext *httputils.SiteContext) *settings.SettingMap {
	if s.Site {
		return siteContext.GetSettings().SiteSettingsMap
	}
	return siteContext.GetSettings().IdentitySettingsMap
}

// ////
// Get Model
// /////
// GetRow is the level's settings, with those it inherits; settings are not rows, so the model is its own row
func (s *SettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Settings, error) {
	if s.Site && siteContext.IdentityModel.Role != consts.IdentityRoleAdmin {
		return nil, apiutils.ErrForbidden
	}
	combined, units := siteContext.GetSettings(), siteContext.GetSettings().GetUnits()
	model := &sitemodels.Settings{Inherited: map[string]string{}}
	fields := model.Fields()
	for _, name := range settingNames {
		if setting := s.level(siteContext).GetSettingAnyTypeByName(settingID(name)); setting != nil {
			*fields[name] = new(showSetting(setting, units))
		}
		inherited := combined.DefaultSettingsMap.GetSettingAnyTypeByName(settingID(name))
		if site := combined.SiteSettingsMap.GetSettingAnyTypeByName(settingID(name)); !s.Site && site != nil {
			inherited = site
		}
		model.Inherited[name] = showSetting(inherited, units)
	}

	// An admin's settings page shows the site's settings beside their own
	if !s.Site && siteContext.IdentityModel.Role == consts.IdentityRoleAdmin {
		site, err := (&SettingsCollection{Site: true}).GetRow(ctx, collectionTx, id, siteContext)
		if err != nil {
			return nil, err
		}
		model.Site = site
	}
	return model, nil
}

func (s *SettingsCollection) ToModel(row *sitemodels.Settings, _ *httputils.SiteContext) *sitemodels.Settings {
	return row
}

func (s *SettingsCollection) GetHTMXResponse(model *sitemodels.Settings) templ.Component {
	if s.Site {
		return site_settings_templ.SettingsPage(model, nil)
	}
	return site_settings_templ.SettingsPage(model.Site, model)
}

// ///
// Update
// ///

// UpdateRow sets the values sent. An empty value clears it, so the level above applies; one not sent is left as it is.
func (s *SettingsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.Settings, model *sitemodels.Settings) (*sitemodels.Settings, error) {
	level, subject := s.level(siteContext), ptr("identity")
	if s.Site {
		subject = ptr("site")
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
		params := &controldb.UpdateSiteIdentitySettingsParams{ID: siteContext.IdentityModel.ID, SiteID: siteContext.SiteID}
		settings.SetSettings(params, level)
		err = q.UpdateSiteIdentitySettings(ctx, *params)
	}
	if err != nil {
		return nil, err
	}
	return s.GetRow(ctx, collectionTx, "", siteContext)
}

func (s *SettingsCollection) UpdateHTMXResponse(model *sitemodels.Settings) templ.Component {
	return site_settings_templ.SettingsSaved()
}
