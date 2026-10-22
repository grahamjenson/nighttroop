package sitedata

import (
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
)

// The handlers check a store can do what its routes need when the router is made; this checks it as it compiles
var (
	_ readWriteRows[sitedb.Activity, sitemodels.SiteActivity]                   = &ActivitiesStore{}
	_ readWriteRows[sitedb.ComponentEvent, sitemodels.SiteComponentEvent]       = &ComponentEventsStore{}
	_ readWriteRows[sitedb.FoodIntake, sitemodels.SiteFoodIntake]               = &FoodIntakesStore{}
	_ readWriteRows[sitedb.GlucoseReading, sitemodels.SiteGlucoseReading]       = &GlucoseReadingsStore{}
	_ readWriteRows[sitedb.InsulinDose, sitemodels.SiteInsulinDose]             = &InsulinDosesStore{}
	_ readWriteRows[sitedb.NotificationEvent, sitemodels.SiteNotificationEvent] = &NotificationEventsStore{}
	_ readWriteRows[sitedb.ProfileSwitch, sitemodels.SiteProfileSwitch]         = &ProfileSwitchesStore{}
	_ readWriteRows[sitedb.TempBasal, sitemodels.SiteTempBasal]                 = &TempBasalsStore{}
	_ readWriteRows[sitedb.TempTarget, sitemodels.SiteTempTarget]               = &TempTargetsStore{}
	// Device statuses and event groups are only read and deleted on the site
	_ readDeleteRows[sitedb.DeviceReport, sitemodels.SiteDeviceStatus]    = &DeviceStatusesStore{}
	_ readDeleteRows[sitedb.EventGroupsBundle, sitemodels.SiteEventGroup] = &EventGroupsStore{}
)

type readDeleteRows[R, M any] interface {
	siteapiutils.RowStore[R, M, *sitedb.Queries]
	siteapiutils.RowGetter[M]
	siteapiutils.RowSearcher[R, M, *sitedb.Queries]
	siteapiutils.RowDeleter[R, M, *sitedb.Queries]
}

type readWriteRows[R, M any] interface {
	readDeleteRows[R, M]
	siteapiutils.RowCreator[R, M, *sitedb.Queries]
	siteapiutils.RowUpdater[R, M, *sitedb.Queries]
}
