package sitecollections

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// UsageHandler is the site's Usage page, for its admins: its storage (usageStorage) and Download Data, or as JSON its
// DB's size and limit
func UsageHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), siteapiutils.ActionGet)
	if siteContext == nil {
		return
	}
	sizeKB, limitKB := siteSize(r)
	WriteResponse(w, r, siteContext, http.StatusOK, map[string]*int64{"size_kb": sizeKB, "limit_kb": limitKB},
		site_settings_templ.Usage(usageStorage(r, sizeKB, limitKB)), true)
}

// usageStorage is the site's storage on its Usage page, its DB's size against its limit and each table's share, as
// the control site shows it, with Truncate. What cannot be measured is left out, as the page is still shown without it.
func usageStorage(r *http.Request, sizeKB, limitKB *int64) templ.Component {
	siteContext := httputils.GetSiteContext(r.Context())
	sizes, err := siteContext.ControlDB.SiteTableSizes(r.Context(), siteContext.SiteID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Cannot measure the site's DB tables", "error", err)
	}
	tables := []*controlmodels.SiteTable{}
	for _, size := range sizes {
		tables = append(tables, &controlmodels.SiteTable{Name: size.Name, Bytes: size.Bytes, Records: size.Records})
	}
	return site_settings_templ.UsageStorage(sizeKB, limitKB, tables)
}

// siteSize is the site's DB size measured now, or as last measured when it cannot be now, and its limit
func siteSize(r *http.Request) (*int64, *int64) {
	siteContext := httputils.GetSiteContext(r.Context())
	if err := siteContext.ControlDB.RecordSiteSize(r.Context(), siteContext.SiteID); err != nil {
		slog.ErrorContext(r.Context(), "Cannot measure the site's DB size", "error", err)
	}
	site, err := siteContext.ControlDB.Queries.GetSite(r.Context(), controldb.GetSiteParams{ID: siteContext.SiteID})
	if err != nil {
		slog.ErrorContext(r.Context(), "Cannot get the site's DB size", "error", err)
		return nil, nil
	}
	return site.EstimatedDbSizeKb, site.MaxDbSizeKbLimit
}

// TruncateHandler truncates the site for its admin from its Usage page, as the control site's Truncate does, recorded as
// the same event, then shows its storage again with a toast of its new size. It rewrites the whole DB, so it shares
// the site's few downloads a minute.
func TruncateHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), siteapiutils.ActionUpdate)
	if siteContext == nil {
		return
	}
	if allowed, wait := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SiteDownloadLimiter(siteContext.SiteID)); !allowed {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(wait))
		return
	}
	started := time.Now()
	beforeKB, limitKB := siteSize(r)
	afterKB, err := siteContext.ControlDB.TruncateSite(r.Context(), siteContext.SiteID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Cannot truncate the site", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	siteContext.ControlDB.Logs.SiteTruncatedEvent(r.Context(), siteContext.SiteID, siteContext.IdentityModel.Name, started, beforeKB, afterKB)
	WriteResponse(w, r, siteContext, http.StatusOK, map[string]*int64{"size_kb": afterKB},
		templ.Join(usageStorage(r, afterKB, limitKB), site_settings_templ.UsageTruncated(afterKB)), false)
}
