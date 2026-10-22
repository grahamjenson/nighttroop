package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitedata"
	"nighttroop.t1d.nz/server/siteapi/sitelogin"
	"nighttroop.t1d.nz/server/siteapi/sitesettings"
	"nighttroop.t1d.nz/server/siteapi/sitesummary"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	//STORES
	syncSettingsStore := siteapiutils.ControlRows[sitemodels.SyncSource, sitemodels.SyncSource](&sitesettings.SyncSettingsStore{})
	identitiesStore := siteapiutils.ControlRows[sitemodels.UserIdentity, sitemodels.UserIdentity](&sitesettings.IdentitiesStore{})
	glucoseReadingsStore := siteapiutils.SiteRows[sitedb.GlucoseReading, sitemodels.SiteGlucoseReading](&sitedata.GlucoseReadingsStore{})
	tempTargetsStore := siteapiutils.SiteRows[sitedb.TempTarget, sitemodels.SiteTempTarget](&sitedata.TempTargetsStore{})
	tempBasalsStore := siteapiutils.SiteRows[sitedb.TempBasal, sitemodels.SiteTempBasal](&sitedata.TempBasalsStore{})
	activitiesStore := siteapiutils.SiteRows[sitedb.Activity, sitemodels.SiteActivity](&sitedata.ActivitiesStore{})
	componentEventsStore := siteapiutils.SiteRows[sitedb.ComponentEvent, sitemodels.SiteComponentEvent](&sitedata.ComponentEventsStore{})
	notificationEventsStore := siteapiutils.SiteRows[sitedb.NotificationEvent, sitemodels.SiteNotificationEvent](&sitedata.NotificationEventsStore{})
	profileSwitchesStore := siteapiutils.SiteRows[sitedb.ProfileSwitch, sitemodels.SiteProfileSwitch](&sitedata.ProfileSwitchesStore{})
	eventGroupsStore := siteapiutils.SiteRows[sitedb.EventGroupsBundle, sitemodels.SiteEventGroup](&sitedata.EventGroupsStore{})
	insulinDosesStore := siteapiutils.SiteRows[sitedb.InsulinDose, sitemodels.SiteInsulinDose](&sitedata.InsulinDosesStore{})
	foodIntakesStore := siteapiutils.SiteRows[sitedb.FoodIntake, sitemodels.SiteFoodIntake](&sitedata.FoodIntakesStore{})
	deviceStatusesStore := siteapiutils.SiteRows[sitedb.DeviceReport, sitemodels.SiteDeviceStatus](&sitedata.DeviceStatusesStore{})
	siteSettingsStore := siteapiutils.ControlRows[sitemodels.Settings, sitemodels.Settings](&sitesettings.SettingsStore{Site: true})
	identitySettingsStore := siteapiutils.ControlRows[sitemodels.Settings, sitemodels.Settings](&sitesettings.SettingsStore{})

	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)

	// Rate Limiter is before auth because auth can be expencive
	r.Use(httputils.SiteRateLimitMiddleware(config))

	static.AddStaticRoutes(r)

	// Publically available (if the person knows the subdomain) endpoints
	r.Group(func(r chi.Router) {
		r.Get("/login", sitelogin.GETLoginPageHandler())
		r.Post("/login", sitelogin.POSTLogin(controlDB, config))
		r.Post("/login/totp", sitelogin.POSTLoginTOTP(controlDB, config))
	})

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(siteapiutils.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(httputils.SiteIdentityRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags

		// Summary page
		r.Get("/", sitesummary.GetSummaryHandler(controlDB, config))
		r.Get("/summary", sitesummary.GetSummaryHandler(controlDB, config))

		// Data Glucose Readings
		r.Get("/data/glucose", siteapiutils.SearchHandler(controlDB, config, glucoseReadingsStore))
		r.Get("/data/glucose/{id}", siteapiutils.GetHandler(controlDB, config, glucoseReadingsStore))
		r.Post("/data/glucose", siteapiutils.CreateHandler(controlDB, config, glucoseReadingsStore))
		r.Post("/data/glucose/{id}", siteapiutils.UpdateHandler(controlDB, config, glucoseReadingsStore))
		r.Delete("/data/glucose/{id}", siteapiutils.DeleteHandler(controlDB, config, glucoseReadingsStore))

		// Data Temp Targets
		r.Get("/data/temp-targets", siteapiutils.SearchHandler(controlDB, config, tempTargetsStore))
		r.Get("/data/temp-targets/{id}", siteapiutils.GetHandler(controlDB, config, tempTargetsStore))
		r.Post("/data/temp-targets", siteapiutils.CreateHandler(controlDB, config, tempTargetsStore))
		r.Post("/data/temp-targets/{id}", siteapiutils.UpdateHandler(controlDB, config, tempTargetsStore))
		r.Delete("/data/temp-targets/{id}", siteapiutils.DeleteHandler(controlDB, config, tempTargetsStore))

		// Data Temp Basals
		r.Get("/data/temp-basals", siteapiutils.SearchHandler(controlDB, config, tempBasalsStore))
		r.Get("/data/temp-basals/{id}", siteapiutils.GetHandler(controlDB, config, tempBasalsStore))
		r.Post("/data/temp-basals", siteapiutils.CreateHandler(controlDB, config, tempBasalsStore))
		r.Post("/data/temp-basals/{id}", siteapiutils.UpdateHandler(controlDB, config, tempBasalsStore))
		r.Delete("/data/temp-basals/{id}", siteapiutils.DeleteHandler(controlDB, config, tempBasalsStore))

		// Data Activities
		r.Get("/data/activities", siteapiutils.SearchHandler(controlDB, config, activitiesStore))
		r.Get("/data/activities/{id}", siteapiutils.GetHandler(controlDB, config, activitiesStore))
		r.Post("/data/activities", siteapiutils.CreateHandler(controlDB, config, activitiesStore))
		r.Post("/data/activities/{id}", siteapiutils.UpdateHandler(controlDB, config, activitiesStore))
		r.Delete("/data/activities/{id}", siteapiutils.DeleteHandler(controlDB, config, activitiesStore))

		// Data Component Events
		r.Get("/data/component-events", siteapiutils.SearchHandler(controlDB, config, componentEventsStore))
		r.Get("/data/component-events/{id}", siteapiutils.GetHandler(controlDB, config, componentEventsStore))
		r.Post("/data/component-events", siteapiutils.CreateHandler(controlDB, config, componentEventsStore))
		r.Post("/data/component-events/{id}", siteapiutils.UpdateHandler(controlDB, config, componentEventsStore))
		r.Delete("/data/component-events/{id}", siteapiutils.DeleteHandler(controlDB, config, componentEventsStore))

		// Data Notifications
		r.Get("/data/notifications", siteapiutils.SearchHandler(controlDB, config, notificationEventsStore))
		r.Get("/data/notifications/{id}", siteapiutils.GetHandler(controlDB, config, notificationEventsStore))
		r.Post("/data/notifications", siteapiutils.CreateHandler(controlDB, config, notificationEventsStore))
		r.Post("/data/notifications/{id}", siteapiutils.UpdateHandler(controlDB, config, notificationEventsStore))
		r.Delete("/data/notifications/{id}", siteapiutils.DeleteHandler(controlDB, config, notificationEventsStore))

		// Data Profile Switches
		r.Get("/data/profile-switches", siteapiutils.SearchHandler(controlDB, config, profileSwitchesStore))
		r.Get("/data/profile-switches/{id}", siteapiutils.GetHandler(controlDB, config, profileSwitchesStore))
		r.Post("/data/profile-switches", siteapiutils.CreateHandler(controlDB, config, profileSwitchesStore))
		r.Post("/data/profile-switches/{id}", siteapiutils.UpdateHandler(controlDB, config, profileSwitchesStore))
		r.Delete("/data/profile-switches/{id}", siteapiutils.DeleteHandler(controlDB, config, profileSwitchesStore))

		// Data Event Groups, which the Nightscout treatments API creates and changes
		r.Get("/data/event-groups", siteapiutils.SearchHandler(controlDB, config, eventGroupsStore))
		r.Get("/data/event-groups/{id}", siteapiutils.GetHandler(controlDB, config, eventGroupsStore))
		r.Delete("/data/event-groups/{id}", siteapiutils.DeleteHandler(controlDB, config, eventGroupsStore))

		// Data Insulin Doses
		r.Get("/data/insulin", siteapiutils.SearchHandler(controlDB, config, insulinDosesStore))
		r.Get("/data/insulin/{id}", siteapiutils.GetHandler(controlDB, config, insulinDosesStore))
		r.Post("/data/insulin", siteapiutils.CreateHandler(controlDB, config, insulinDosesStore))
		r.Post("/data/insulin/{id}", siteapiutils.UpdateHandler(controlDB, config, insulinDosesStore))
		r.Delete("/data/insulin/{id}", siteapiutils.DeleteHandler(controlDB, config, insulinDosesStore))

		// Data Food Intakes
		r.Get("/data/food", siteapiutils.SearchHandler(controlDB, config, foodIntakesStore))
		r.Get("/data/food/{id}", siteapiutils.GetHandler(controlDB, config, foodIntakesStore))
		r.Post("/data/food", siteapiutils.CreateHandler(controlDB, config, foodIntakesStore))
		r.Post("/data/food/{id}", siteapiutils.UpdateHandler(controlDB, config, foodIntakesStore))
		r.Delete("/data/food/{id}", siteapiutils.DeleteHandler(controlDB, config, foodIntakesStore))

		// Data Device Statuses
		r.Get("/data/device-status", siteapiutils.SearchHandler(controlDB, config, deviceStatusesStore))
		r.Get("/data/device-status/{id}", siteapiutils.GetHandler(controlDB, config, deviceStatusesStore))
		r.Delete("/data/device-status/{id}", siteapiutils.DeleteHandler(controlDB, config, deviceStatusesStore))

		// Settings Sync
		r.Get("/settings/sync", siteapiutils.SearchHandler(controlDB, config, syncSettingsStore))
		r.Post("/settings/sync/{id}", siteapiutils.UpsertHandler(controlDB, config, syncSettingsStore))
		r.Delete("/settings/sync/{id}", siteapiutils.DeleteHandler(controlDB, config, syncSettingsStore))

		// Settings Identities
		r.Get("/settings/identities", siteapiutils.SearchHandler(controlDB, config, identitiesStore))
		r.Get("/settings/identities/{id}", siteapiutils.GetHandler(controlDB, config, identitiesStore))
		r.Post("/settings/identities", siteapiutils.CreateHandler(controlDB, config, identitiesStore))
		r.Post("/settings/identities/{id}", siteapiutils.UpdateHandler(controlDB, config, identitiesStore))
		r.Delete("/settings/identities/{id}", siteapiutils.DeleteHandler(controlDB, config, identitiesStore))

		// Settings for the site (admins only) and the signed in identity
		// One page shows the site's settings beside an admin's own, or just the identity's
		r.Get("/settings", siteapiutils.GetHandler(controlDB, config, identitySettingsStore))
		r.Get("/settings/site", siteapiutils.GetHandler(controlDB, config, siteSettingsStore))
		r.Post("/settings/site", siteapiutils.UpdateHandler(controlDB, config, siteSettingsStore))
		r.Get("/settings/identity", siteapiutils.GetHandler(controlDB, config, identitySettingsStore))
		r.Post("/settings/identity", siteapiutils.UpdateHandler(controlDB, config, identitySettingsStore))

		// Settings Settings
		r.Get("/settings/config", sitesettings.GETConfigSettingsHandler(controlDB, config))
		r.Post("/settings/config", sitesettings.POSTConfigSettingsHandler(controlDB, config))

		// Actions
		r.Post("/logout", sitelogin.POSTLogoutHandler(config))

	})

	return r
}
