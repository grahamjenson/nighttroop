package collections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileCollection struct {
	converter func(c *site.ProfilePreset) nsutils.Record
}

func NewProfileCollection() ProfileCollection {
	return ProfileCollection{converter: nsutils.DBModelToRecord(nsutils.Profile, nightscout.DBModelToProfile)}
}

func (c ProfileCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var profilePresets []*site.ProfilePreset
	var rerr error

	// There are three options,
	if !query.HasFilters {
		profilePresets, rerr = siteContext.SiteDB.Queries.AllProfilePresets(ctx, nsquery.ProfilePresetsAll(query))
	} else {
		profilePresets, rerr = siteContext.SiteDB.Queries.SearchProfilePresets(ctx, nsquery.ProfilePresetsSearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query profile presets from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(profilePresets, c.converter), nil
}

func (c ProfileCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c ProfileCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return nil, false, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}
