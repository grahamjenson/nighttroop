package sitecollections

import (
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
)

type SiteStore struct {
	collections map[string]SiteCollection
}

func NewSiteStore() *SiteStore {
	collections := map[string]SiteCollection{
		"syncSettings":       &sitesettings.SyncSettingsCollection{},
		"identities":         &sitesettings.IdentitiesCollection{},
		"siteSettings":       &sitesettings.SettingsCollection{Site: true},
		"identitySettings":   &sitesettings.SettingsCollection{Site: false},
		"glucoseReadings":    &sitedata.GlucoseReadingsCollection{},
		"tempTargets":        &sitedata.TempTargetsCollection{},
		"tempBasals":         &sitedata.TempBasalsCollection{},
		"activities":         &sitedata.ActivitiesCollection{},
		"componentEvents":    &sitedata.ComponentEventsCollection{},
		"notificationEvents": &sitedata.NotificationEventsCollection{},
		"profileSwitches":    &sitedata.ProfileSwitchesCollection{},
		"eventGroups":        &sitedata.EventGroupsCollection{},
		"insulinDoses":       &sitedata.InsulinDosesCollection{},
		"foodIntakes":        &sitedata.FoodIntakesCollection{},
		"deviceStatuses":     &sitedata.DeviceStatusesCollection{},
	}
	return &SiteStore{
		collections: collections,
	}
}
