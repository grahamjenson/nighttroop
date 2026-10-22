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
		siteapiutils.Routes(r, "/data/glucose", glucoseReadingsStore)

		// Data Temp Targets
		siteapiutils.Routes(r, "/data/temp-targets", tempTargetsStore)

		// Data Temp Basals
		siteapiutils.Routes(r, "/data/temp-basals", tempBasalsStore)

		// Data Activities
		siteapiutils.Routes(r, "/data/activities", activitiesStore)

		// Data Component Events
		siteapiutils.Routes(r, "/data/component-events", componentEventsStore)

		// Data Notifications
		siteapiutils.Routes(r, "/data/notifications", notificationEventsStore)

		// Data Profile Switches
		siteapiutils.Routes(r, "/data/profile-switches", profileSwitchesStore)

		// Data Event Groups, which the Nightscout treatments API creates and changes
		siteapiutils.Routes(r, "/data/event-groups", eventGroupsStore)

		// Data Insulin Doses
		siteapiutils.Routes(r, "/data/insulin", insulinDosesStore)

		// Data Food Intakes
		siteapiutils.Routes(r, "/data/food", foodIntakesStore)

		// Data Device Statuses
		siteapiutils.Routes(r, "/data/device-status", deviceStatusesStore)

		// Settings Sync
		r.Get("/settings/sync", siteapiutils.SearchHandler(syncSettingsStore))
		r.Post("/settings/sync/{id}", siteapiutils.UpsertHandler(syncSettingsStore))
		r.Delete("/settings/sync/{id}", siteapiutils.DeleteHandler(syncSettingsStore))

		// Settings Identities
		siteapiutils.Routes(r, "/settings/identities", identitiesStore)

		// Settings for the site (admins only) and the signed in identity
		// One page shows the site's settings beside an admin's own, or just the identity's
		r.Get("/settings", siteapiutils.GetHandler(identitySettingsStore))
		r.Get("/settings/site", siteapiutils.GetHandler(siteSettingsStore))
		r.Post("/settings/site", siteapiutils.UpdateHandler(siteSettingsStore))
		r.Get("/settings/identity", siteapiutils.GetHandler(identitySettingsStore))
		r.Post("/settings/identity", siteapiutils.UpdateHandler(identitySettingsStore))

		// Settings Settings
		r.Get("/settings/config", sitesettings.GETConfigSettingsHandler(controlDB, config))
		r.Post("/settings/config", sitesettings.POSTConfigSettingsHandler(controlDB, config))

		// Actions
		r.Post("/logout", sitelogin.POSTLogoutHandler(config))

	})

	return r
}
