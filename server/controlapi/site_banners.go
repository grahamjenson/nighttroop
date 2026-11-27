package controlapi

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_banners_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/types"
)

// siteBannerForm is a banner as an admin makes it: its sites as one text, split at commas and spaces, and how many
// days it lasts
type siteBannerForm struct {
	Message string              `json:"message"`
	Level   string              `json:"level"`
	Sites   string              `json:"sites"`
	Days    types.FlexibleInt64 `json:"days"`
}

// SiteBannersHandler is every site banner, newest first
func SiteBannersHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	rows, err := controlContext.ControlDB.Queries.AllSiteBanners(r.Context())
	banners, herr := apiutils.Searched(rows, err, "site banners", controlmodels.SiteBannerFromDB)
	if herr != nil {
		controlcollections.WriteError(w, r, acceptType, herr)
		return
	}
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, banners, control_banners_templ.Banners(banners))
}

// CreateSiteBannerHandler makes a banner that sites show until it expires. Only admins can, and only so many, each
// with a message, for a time, and for sites that are named as a site is.
func CreateSiteBannerHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	form, herr := apiutils.ExtractModelFromRequest[siteBannerForm](httputils.GetContentType(r), r)
	if herr != nil {
		controlcollections.WriteError(w, r, acceptType, herr)
		return
	}
	message, level := strings.TrimSpace(form.Message), form.Level
	if level == "" {
		level = controlmodels.SiteBannerLevels[0]
	}
	sites := strings.FieldsFunc(form.Sites, func(c rune) bool { return c == ',' || c == ' ' || c == '\n' || c == '\t' })
	slices.Sort(sites)
	sites = slices.Compact(sites)
	if sites == nil {
		sites = []string{}
	}
	var refused string
	switch {
	case message == "" || len(message) > controlmodels.SiteBannerMessageLimit.MaxLength:
		refused = "errors.control.banner_message"
	case !slices.Contains(controlmodels.SiteBannerLevels, level):
		refused = "errors.control.banner_level"
	case form.Days < 1 || form.Days > controlmodels.MaxSiteBannerDays:
		refused = "errors.control.banner_days"
	case len(sites) > controlmodels.MaxSiteBannerSites || slices.ContainsFunc(sites, func(id string) bool { return !utils.SyntacticallyValidateSiteID(id) }):
		refused = "errors.control.banner_sites"
	}
	if refused == "" {
		if count, err := controlContext.ControlDB.Queries.CountSiteBanners(r.Context()); err != nil || count >= controlmodels.MaxSiteBanners {
			refused = "errors.control.banner_most"
		}
	}
	if refused != "" {
		controlcollections.WriteError(w, r, acceptType, apiutils.ErrBadBodyWithHint(i18n.NewError(refused, i18n.Data{"Max": controlmodels.MaxSiteBanners,
			"Days": controlmodels.MaxSiteBannerDays, "Sites": controlmodels.MaxSiteBannerSites, "Length": controlmodels.SiteBannerMessageLimit.MaxLength})))
		return
	}
	siteIDs, _ := json.Marshal(sites)
	row, err := controlContext.ControlDB.Queries.CreateSiteBanner(r.Context(), controldb.CreateSiteBannerParams{Message: message, Level: level,
		SiteIdsJson: string(siteIDs), ExpiresAt: time.Now().AddDate(0, 0, int(form.Days)).UnixMilli(), CreatedByIdentityName: controlContext.IdentityModel.Name})
	if herr := apiutils.ErrorFor(err, "site banners"); herr != nil {
		controlcollections.WriteError(w, r, acceptType, herr)
		return
	}
	banner := controlmodels.SiteBannerFromDB(row)
	controlcollections.WriteResponse(w, r, acceptType, http.StatusCreated, banner, control_banners_templ.Added(banner))
}

// DeleteSiteBannerHandler deletes a banner, so no site shows it. Only admins can.
func DeleteSiteBannerHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var deleted int64
	if err == nil {
		deleted, err = controlContext.ControlDB.Queries.DeleteSiteBanner(r.Context(), controldb.DeleteSiteBannerParams{ID: id})
	}
	if err != nil || deleted == 0 {
		controlcollections.WriteError(w, r, acceptType, apiutils.ErrNotFound)
		return
	}
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, map[string]string{"status": "deleted"}, control_banners_templ.Deleted())
}
