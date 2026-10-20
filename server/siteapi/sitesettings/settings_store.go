package sitesettings

import (
	"context"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
)

// SettingsStore is a level of settings: the site's, which only admins can change,
// or the signed in identity's own, which override the site's
type SettingsStore struct {
	Site bool
}

// settingNames are the settings the pages show so far, in order
var settingNames = []string{"units", "timezone", "language", "date_format", "time_format", "week_start_day"}

// ////
// Shared
// ////
func (s *SettingsStore) GetName() string {
	if s.Site {
		return "site_settings"
	}
	return "identity_settings"
}

func (s *SettingsStore) level(siteContext *httputils.SiteContext) *settings.SettingMap {
	if s.Site {
		return siteContext.GetSettings().SiteSettingsMap
	}
	return siteContext.GetSettings().IdentitySettingsMap
}

// ////
// Get Model
// /////
func (s *SettingsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Settings, *apiutils.HTTPError) {
	if s.Site && siteContext.IdentityModel.Role != consts.IdentityRoleAdmin {
		return nil, apiutils.ErrForbidden
	}
	combined := siteContext.GetSettings()
	model := sitemodels.SettingsFromMap(s.level(siteContext))
	model.Inherited = map[string]string{}
	for _, name := range settingNames {
		inherited := combined.DefaultSettingsMap.GetSettingAnyTypeByName(name)
		if site := combined.SiteSettingsMap.GetSettingAnyTypeByName(name); !s.Site && site != nil {
			inherited = site
		}
		model.Inherited[name] = fmt.Sprint(inherited.Value)
	}
	// An admin's settings page shows the site's settings beside their own
	if !s.Site && siteContext.IdentityModel.Role == consts.IdentityRoleAdmin {
		site, herr := (&SettingsStore{Site: true}).GetModel(ctx, tx, id, siteContext)
		if herr != nil {
			return nil, herr
		}
		model.Site = site
	}
	return model, nil
}

func (s *SettingsStore) GetHTMXResponse(model *sitemodels.Settings) templ.Component {
	if s.Site {
		return site_settings_templ.SettingsPage(model, nil)
	}
	return site_settings_templ.SettingsPage(model.Site, model)
}

// ///
// Update
// ///

// UpdateModel sets the values sent. An empty value clears it, so the level above applies; one not sent is left as it is.
func (s *SettingsStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.Settings, model *sitemodels.Settings) *apiutils.HTTPError {
	level, subject := s.level(siteContext), ptr("identity")
	if s.Site {
		subject = ptr("site")
	}
	sent := map[string]*string{
		"units": model.Units, "timezone": model.Timezone, "language": model.Language,
		"date_format": model.DateFormat, "time_format": model.TimeFormat, "week_start_day": model.WeekStartDay,
	}
	for _, name := range settingNames {
		if sent[name] == nil {
			continue
		}
		var value any
		if *sent[name] != "" {
			value = *sent[name]
		}
		if err := level.SetSettingByName(name, value, nil, subject); err != nil {
			return apiutils.ErrBadBodyWithHint(err.Error())
		}
	}

	q, err := tx.Control()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	// Every setting is written, so the ones not on the page keep their values
	if s.Site {
		params := &controldb.UpdateSitesSettingsParams{ID: siteContext.SiteModel.ID}
		settings.SetSettings(params, level)
		err = q.UpdateSitesSettings(ctx, *params)
	} else {
		params := &controldb.UpdateSiteIdentitySettingsParams{ID: siteContext.IdentityModel.ID, SiteID: siteContext.Subdomain}
		settings.SetSettings(params, level)
		err = q.UpdateSiteIdentitySettings(ctx, *params)
	}
	if err != nil {
		slog.Error("Error updating settings", "error", err, "site", s.Site)
		return apiutils.ErrInternalServerError
	}

	updated, herr := s.GetModel(ctx, tx, id, siteContext)
	if herr != nil {
		return herr
	}
	*model = *updated
	return nil
}

func (s *SettingsStore) UpdateHTMXResponse(model *sitemodels.Settings) templ.Component {
	return site_settings_templ.SettingsSaved()
}
