package sitepresets

import (
	"context"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_presets_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// LoopSettingsCollection is each set of settings the loop apps sent, to see and inspect. They come only from the
// loops, so the site does not make or edit them; deleting one leaves the profiles and device statuses sent with it,
// which are then read back without it.
type LoopSettingsCollection struct{}

func (s *LoopSettingsCollection) GetName() string {
	return "loop_settings"
}

// GetRoute is where the collection is in its URL, /presets/loop-settings
func (s *LoopSettingsCollection) GetRoute() string {
	return "loop-settings"
}

func (s *LoopSettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.LoopSetting, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetLoopSettings, sitedb.GetLoopSettingsParams{ID: id})
}

func (s *LoopSettingsCollection) ToModel(row *sitedb.LoopSetting, siteContext *httputils.SiteContext) *sitemodels.SiteLoopSettings {
	return sitemodels.LoopSettingsFromDB(row)
}

// GetHTMXResponse is the settings, to inspect, in a modal
func (s *LoopSettingsCollection) GetHTMXResponse(model *sitemodels.SiteLoopSettings) templ.Component {
	return site_presets_templ.LoopSettingsModal(model)
}

// SearchRows are the settings, the last seen first
func (s *LoopSettingsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.LoopSetting, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).SearchLoopSettings, sitedb.SearchLoopSettingsParams{Limit: query.Limit, Skip: query.Skip})
}

func (s *LoopSettingsCollection) SearchHTMXResponse(models []*sitemodels.SiteLoopSettings, query *apiutils.Query) templ.Component {
	return site_presets_templ.LoopSettingsTable(models, query.Skip, query.Limit)
}

// DeleteRow deletes the settings for good, leaving the records sent with them
func (s *LoopSettingsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	_, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).DeleteLoopSettings, sitedb.DeleteLoopSettingsParams{ID: id})
	return err
}

func (s *LoopSettingsCollection) DeleteHTMXResponse(model *sitemodels.SiteLoopSettings) templ.Component {
	return site_presets_templ.LoopSettingsDeleted()
}
