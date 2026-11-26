// Package banner is the one place that decides what a page says in the strip at its top, for the control site and
// for a site, and draws it: a service that is not working, a resource running high, a site over its storage, a
// message a control admin wrote for sites. A page asks ForControl or ForSite, and shows what it gets with Strip.
package banner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/display"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
)

// Level is how much a banner matters, which is how it is coloured
type Level string

const (
	Info    Level = "info"
	Warning Level = "warning"
	Error   Level = "error"
)

// levelRanks are how much each level matters, as a number, so a rule can be for every level under one
var levelRanks = map[Level]int{Info: 2, Warning: 5, Error: 8}

// hideFrom is the rank from which a banner cannot be hidden: a warning or an error stays until it is no longer true
const hideFrom = 5

// CanHide is whether a person may hide a banner of this level, which only one that is good to know is
func (l Level) CanHide() bool {
	rank, known := levelRanks[l]
	return known && rank < hideFrom
}

// Banner is one line of the strip: a title in bold, its text, then any links and what follows them
type Banner struct {
	// ID is the line's id on the page
	ID    string
	Level Level
	Title string
	Text  string
	Links []Link
	After string
}

// Link is a link in a banner's line, with the text that follows it
type Link struct {
	ID, Href, Text, After string
}

// s3OffReasons say why the server turned S3 off at start, by config.ServerStatus.S3Off
var s3OffReasons = map[string]string{
	config.S3OffConflict:    "control.status_banner.s3_off_conflict",
	config.S3OffCheck:       "control.status_banner.s3_off_check",
	config.S3OffReplication: "control.status_banner.s3_off_replication",
}

// unavailableReasons are the words for why a site is unavailable
var unavailableReasons = map[string]string{
	db.UnavailableMissing: "control.status_banner.unavailable_missing", db.UnavailableCannotOpen: "control.status_banner.unavailable_cannot_open",
	db.UnavailableMigrationFailed: "control.status_banner.unavailable_migration_failed",
	db.UnavailableSchemaDiffers:   "control.status_banner.unavailable_schema_differs",
}

// maxUnavailableShown is how many unavailable sites a banner names, so it stays short when S3 is down for every site
const maxUnavailableShown = 10

// resourceTitles are the ids of each resource's title
var resourceTitles = map[string]string{
	"memory": "control.status_banner.memory_high", "disk": "control.status_banner.disk_high", "cpu": "control.status_banner.cpu_high",
}

// ForControl is what the control site's pages warn its admins of: optional services that are not working, resources
// running high, sites near their limit, and sites that cannot be served
func ForControl(ctx context.Context, status config.ServerStatus) []Banner {
	var banners []Banner
	add := func(level Level, name, title, text string) {
		banners = append(banners, Banner{ID: "control--status-banner--" + name, Level: level, Title: title, Text: text})
	}
	switch reason, off := s3OffReasons[status.S3Off]; {
	case status.S3Down:
		add(Error, "s3-down", i18n.T(ctx, "control.status_banner.s3_down_title"), i18n.T(ctx, "control.status_banner.s3_down_text"))
	case off:
		add(Error, "s3-off", i18n.T(ctx, "control.status_banner.s3_off_title"), i18n.T(ctx, reason)+" "+i18n.T(ctx, "control.status_banner.s3_off_text"))
	case !status.S3Working:
		add(Warning, "s3", i18n.T(ctx, "control.status_banner.s3_title"), i18n.T(ctx, "control.status_banner.s3_text"))
	}
	if !status.SentryWorking {
		add(Warning, "sentry", i18n.T(ctx, "control.status_banner.sentry_title"), i18n.T(ctx, "control.status_banner.sentry_text"))
	}
	if !status.PagerDutyWorking {
		add(Warning, "pagerduty", i18n.T(ctx, "control.status_banner.pagerduty_title"), i18n.T(ctx, "control.status_banner.pagerduty_text"))
	}
	for _, level := range status.HighResources {
		add(Warning, level.Resource, i18n.T(ctx, resourceTitles[level.Resource])+":", resourceWarning(ctx, level))
	}
	if status.SitesNearLimit() {
		title := i18n.T(ctx, "control.status_banner.sites_near_limit")
		if status.SitesAtLimit() {
			title = i18n.T(ctx, "control.status_banner.sites_at_limit")
		}
		add(Warning, "sites", title+":", i18n.T(ctx, "control.status_banner.sites_text", i18n.Data{"SiteCount": status.Sites, "MaxSites": status.MaxSites}))
	}
	if len(status.UnavailableSites) > 0 {
		unavailable := Banner{ID: "control--status-banner--unavailable", Level: Error,
			Title: i18n.T(ctx, "control.status_banner.unavailable_title", i18n.Data{"SiteCount": len(status.UnavailableSites)})}
		siteIDs := slices.Sorted(maps.Keys(status.UnavailableSites))
		shown := siteIDs[:min(len(siteIDs), maxUnavailableShown)]
		for i, siteID := range shown {
			after := " (" + i18n.T(ctx, unavailableReasons[status.UnavailableSites[siteID]]) + ")"
			if i < len(shown)-1 {
				after += ","
			}
			unavailable.Links = append(unavailable.Links, Link{ID: "control--status-banner--unavailable--" + siteID, Href: "/sites/" + siteID, Text: siteID, After: after})
		}
		if more := len(siteIDs) - len(shown); more > 0 {
			unavailable.After = i18n.T(ctx, "control.status_banner.unavailable_more", i18n.Data{"SiteCount": more})
		}
		banners = append(banners, unavailable)
	}
	return banners
}

// resourceWarning says how high a resource was at the last check: the CPU's seconds at 100%, the memory used and the
// disk free, in bytes where its total is known
func resourceWarning(ctx context.Context, level config.ResourceLevel) string {
	value, limit := fmt.Sprintf("%.0f", level.Value), fmt.Sprintf("%.0f", level.Limit)
	switch {
	case level.Resource == "cpu":
		return i18n.T(ctx, "control.status_banner.cpu_warning", i18n.Data{"CPUSeconds": value, "WarnAt": limit})
	case level.Total == nil:
		return i18n.T(ctx, "control.status_banner.percent_warning", i18n.Data{"UsedPercent": value, "WarnAt": limit})
	case level.Resource == "disk":
		free := *level.Total * (100 - level.Value) / 100
		return i18n.T(ctx, "control.status_banner.disk_warning", i18n.Data{"FreeBytes": display.Bytes(free), "TotalBytes": display.Bytes(*level.Total)})
	}
	return i18n.T(ctx, "control.status_banner.memory_warning", i18n.Data{"UsedBytes": display.Bytes(*level.Total*level.Value/100), "TotalBytes": display.Bytes(*level.Total)})
}

// nearStoragePercent is how much of its storage limit a site holds before its pages say it is near it
const nearStoragePercent = 90

// ForSite is what a site's pages tell whoever is signed in to it: that nothing more is stored, that the site is over
// its storage, and the messages a control admin wrote for the site or for every site
func ForSite(ctx context.Context) []Banner {
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil || siteContext.SiteModel == nil {
		return nil
	}
	var written []*controldb.SiteBanner
	if siteContext.ControlDB != nil {
		var err error
		if written, err = siteContext.ControlDB.Queries.ActiveSiteBanners(ctx, controldb.ActiveSiteBannersParams{Now: time.Now().UnixMilli()}); err != nil {
			slog.ErrorContext(ctx, "cannot read the site banners", "error", err)
		}
	}
	return OfSite(ctx, siteContext.SiteModel, httputils.SiteWriteRefused(siteContext), written)
}

// OfSite is ForSite's banners for a site, from why its writes are refused, "" when they are not, and the banners
// control admins wrote that have not expired
func OfSite(ctx context.Context, site *controldb.Site, writesRefused string, written []*controldb.SiteBanner) []Banner {
	var banners []Banner
	switch size, limit := site.EstimatedDbSizeKb, site.MaxDbSizeKbLimit; {
	case writesRefused != "":
		banners = append(banners, Banner{ID: "site--banner--storage-full", Level: Error, Title: i18n.T(ctx, "site.banner.storage_full"), Text: writesRefused})
	case size != nil && limit != nil && *size > *limit:
		banners = append(banners, Banner{ID: "site--banner--over-storage", Level: Warning, Title: i18n.T(ctx, "site.banner.over_storage"),
			Text: i18n.T(ctx, "site.banner.over_storage_text", i18n.Data{"Size": display.Bytes(float64(*size) * 1024), "Limit": display.Bytes(float64(*limit) * 1024)})})
	case size != nil && limit != nil && *size*100 >= *limit*nearStoragePercent:
		banners = append(banners, Banner{ID: "site--banner--near-storage", Level: Info, Title: i18n.T(ctx, "site.banner.near_storage"),
			Text: i18n.T(ctx, "site.banner.near_storage_text", i18n.Data{"Size": display.Bytes(float64(*size) * 1024), "Limit": display.Bytes(float64(*limit) * 1024)})})
	}
	for _, row := range written {
		var siteIDs []string
		if err := json.Unmarshal([]byte(row.SiteIdsJson), &siteIDs); err != nil || (len(siteIDs) > 0 && !slices.Contains(siteIDs, site.ID)) {
			continue
		}
		level := Info
		if row.Level == string(Warning) {
			level = Warning
		}
		banners = append(banners, Banner{ID: fmt.Sprintf("site--banner--%d", row.ID), Level: level, Text: row.Message})
	}
	return banners
}
