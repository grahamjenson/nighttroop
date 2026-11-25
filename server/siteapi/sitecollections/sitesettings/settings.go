package sitesettings

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/i18n"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/display"
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
var settingNames = []string{"units", "timezone", "language", "date_format", "time_format", "week_start_day", "number_format",
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
		return display.GlucoseValue(units.FromMGDL(float64(mgdl)), units)
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

// editing are the settings the collection reads and changes, and which of them is its level: the site's over the
// defaults, the signed in identity's over the site's, or for Other another identity's over the site's. They are a
// copy, so the request's own settings change only once a change is saved.
func (s *SettingsCollection) editing(siteContext *httputils.SiteContext, identity *controldb.SiteIdentity) (*settings.ContextSettings, *settings.Settings) {
	editing := *siteContext.GetOwnSettings()
	if s.Site {
		editing.Identity = settings.Settings{}
		return &editing, &editing.Site
	}
	if s.Other {
		editing.Identity, editing.IdentitySince = settings.SettingsFromJSON(identity.SettingsJson), identity.CreatedAt
	}
	return &editing, &editing.Identity
}

// glucoseUnits are the units a glucose setting is shown and typed in, as its label says: another identity's are in its
// own units, or the site's it follows, and the site's and the signed in identity's own are in the signed in identity's
func (s *SettingsCollection) glucoseUnits(siteContext *httputils.SiteContext, editing *settings.ContextSettings) *consts.Units {
	if s.Other {
		return new(editing.GetUnits())
	}
	return new(siteContext.GetSettings().GetUnits())
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
	return siteapiutils.ControlQuery(ctx, collectionTx, (*controldb.Queries).GetSiteIdentityByID, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.SiteID, ID: identityID})
}

// ////
// Get Model
// /////
// GetRow is the level's settings, with those it inherits; settings are not rows, so the model is its own row
func (s *SettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Settings, error) {
	if s.Site && !can.IdentityReadSiteSettings(siteContext.IdentityModel.Role) {
		return nil, apiutils.ErrForbidden
	}
	identity, err := s.identity(ctx, collectionTx, id, siteContext)
	if err != nil {
		return nil, err
	}
	editing, level := s.editing(siteContext, identity)
	units := s.glucoseUnits(siteContext, editing)
	model := &sitemodels.Settings{Inherited: map[string]string{}}
	if !s.Site {
		model.OwnerID = &identity.ID
	}
	if s.Other {
		model.IdentityID, model.IdentityName = &identity.ID, &identity.Name
	}
	fields := model.Fields()
	// What the level inherits: the defaults for the site's, and the site's over the defaults for an identity's
	below := &settings.ContextSettings{DefaultSettings: editing.DefaultSettings, SiteSince: editing.SiteSince}
	if !s.Site {
		below.Site = editing.Site
	}
	for _, name := range settingNames {
		if setting := editing.Own(level, settingID(name)); setting != nil {
			*fields[name] = new(showSetting(setting, units))
		}
		if inherited, err := below.Get(settingID(name)); err == nil {
			model.Inherited[name] = showSetting(inherited, units)
		}
	}

	// An admin's settings page shows the site's settings beside their own
	if !s.Site && !s.Other && can.IdentityReadSiteSettings(siteContext.IdentityModel.Role) {
		site, err := (&SettingsCollection{Site: true}).GetRow(ctx, collectionTx, id, siteContext)
		if err != nil {
			return nil, err
		}
		model.Site = site
	}
	return model, nil
}

// Watches is whether a settings change is streamed to the identity: every level is the same model, so its own
// settings and alarms are only its own changes and the site's, which it inherits and an admin's are shown beside.
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
		return true
	}
	return *changed.OwnerID == siteContext.IdentityModel.ID
}

// Hides is whether a settings change is sent as {}: the site's, to a non-admin, who cannot read them but whose
// inherited settings have changed
func (s *SettingsCollection) Hides(event events.Event, siteContext *httputils.SiteContext) bool {
	changed, ok := event.Model.(*sitemodels.Settings)
	return ok && changed.OwnerID == nil && !can.IdentityReadSiteSettings(siteContext.IdentityModel.Role)
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
	editing, level := s.editing(siteContext, identity)
	if err := s.setSent(siteContext, editing, level, model); err != nil {
		return nil, err
	}
	// The settings in effect must make sense together
	if err := editing.Consistent(); err != nil {
		return nil, apiutils.ErrBadBodyWithHint(err)
	}
	if err := s.save(ctx, collectionTx, siteContext, identity, level); err != nil {
		return nil, err
	}
	s.logSiteChanges(ctx, collectionTx, siteContext, editing)
	// The request's own settings follow what was saved, for the settings it answers with
	switch own := siteContext.GetOwnSettings(); {
	case s.Site:
		own.Site = *level
	case !s.Other:
		own.Identity = *level
	}
	return s.GetRow(ctx, collectionTx, id, siteContext)
}

// setSent sets the level's settings to the values the model has, clearing those sent empty
func (s *SettingsCollection) setSent(siteContext *httputils.SiteContext, editing *settings.ContextSettings, level *settings.Settings, model *sitemodels.Settings) error {
	// Units before any change sent, as the glucose sent was typed in those the page showed
	sent, units := model.Fields(), s.glucoseUnits(siteContext, editing)
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
				return apiutils.ErrBadBodyWithHint(i18n.NewError("errors.settings.number", i18n.Data{"Name": name}))
			}
			value = int64(math.Round(units.ToMGDL(glucose)))
		}
		if err := level.Set(settingID(name), value); err != nil {
			return apiutils.ErrBadBodyWithHint(err)
		}
	}
	return nil
}

// save writes every setting of the level, so the ones not on the page keep their values
func (s *SettingsCollection) save(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, identity *controldb.SiteIdentity, level *settings.Settings) error {
	q, err := collectionTx.Control()
	if err != nil {
		return err
	}
	stored, err := level.JSON()
	if err != nil {
		return err
	}
	if s.Site {
		return q.UpdateSiteSettings(ctx, controldb.UpdateSiteSettingsParams{ID: siteContext.SiteModel.ID, SettingsJson: stored})
	}
	return q.UpdateSiteIdentitySettings(ctx, controldb.UpdateSiteIdentitySettingsParams{ID: identity.ID, SiteID: siteContext.SiteID, SettingsJson: stored})
}

// logSiteChanges logs a new site timezone or week start day as an event, once saved, so the background jobs that
// follow it, such as the site's days, see it
func (s *SettingsCollection) logSiteChanges(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, saved *settings.ContextSettings) {
	if !s.Site {
		return
	}
	if was, now := siteContext.GetSiteTimezone().String(), saved.Location().String(); now != was {
		collectionTx.OnCommit(func() {
			siteContext.ControlDB.Logs.SiteTimezoneChangedEvent(ctx, siteContext.SiteID, siteContext.IdentityModel.Name, was, now)
		})
	}
	if was, now := siteContext.GetSiteWeekStartDay(), saved.GetWeekStartDay(); now != was {
		collectionTx.OnCommit(func() {
			siteContext.ControlDB.Logs.SiteWeekStartChangedEvent(ctx, siteContext.SiteID, siteContext.IdentityModel.Name, was, now)
		})
	}
}

func (s *SettingsCollection) UpdateHTMXResponse(model *sitemodels.Settings) templ.Component {
	return site_settings_templ.SettingsSaved()
}
