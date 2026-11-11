package sitepresets

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

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

// GetRow is the settings, with what changed from the set of the same app last seen before them
func (s *LoopSettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteLoopSettings, error) {
	row, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetLoopSettings, sitedb.GetLoopSettingsParams{ID: id})
	if err != nil {
		return nil, err
	}
	previous, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetPreviousLoopSettings, sitedb.GetPreviousLoopSettingsParams{App: row.App, LastSeenAt: row.LastSeenAt, ID: row.ID})
	model := s.fromDB(row, siteContext)
	switch {
	case errors.Is(err, sql.ErrNoRows): // the app's first set
	case err != nil:
		return nil, err
	default:
		model.PreviousID = &previous.ID
		before := sitemodels.LoopSettingsFor(siteContext.IdentityModel, json.RawMessage(previous.SettingsJson))
		model.Changes = sitemodels.LoopSettingsChanges(row.App, before, model.Settings, model.Units)
	}
	return model, nil
}

// fromDB is the model of the settings as the identity may see them, with the site's units, which glucose in their
// changes is in
func (s *LoopSettingsCollection) fromDB(row *sitedb.LoopSetting, siteContext *httputils.SiteContext) *sitemodels.SiteLoopSettings {
	model := sitemodels.LoopSettingsFromDB(row)
	model.Settings = sitemodels.LoopSettingsFor(siteContext.IdentityModel, model.Settings)
	model.Units = siteContext.GetSettings().GetUnits()
	return model
}

// ToModel is the settings as read, which are already the model
func (s *LoopSettingsCollection) ToModel(row *sitemodels.SiteLoopSettings, siteContext *httputils.SiteContext) *sitemodels.SiteLoopSettings {
	return row
}

// GetHTMXResponse is the settings, to inspect, in a modal
func (s *LoopSettingsCollection) GetHTMXResponse(model *sitemodels.SiteLoopSettings) templ.Component {
	return site_presets_templ.LoopSettingsModal(model)
}

// SearchRows are the settings, the last seen first
func (s *LoopSettingsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteLoopSettings, error) {
	rows, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).SearchLoopSettings, sitedb.SearchLoopSettingsParams{Limit: query.Limit, Skip: query.Skip})
	models := make([]*sitemodels.SiteLoopSettings, 0, len(rows))
	for _, row := range rows {
		models = append(models, s.fromDB(row, siteContext))
	}
	return models, err
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
