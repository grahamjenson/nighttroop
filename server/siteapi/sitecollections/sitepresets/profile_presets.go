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

// ProfilePresetsCollection is the profiles loops upload in profile stores. The site shows, renames and deletes
// them; making and editing their schedules stays with the loops.
type ProfilePresetsCollection struct{}

func (s *ProfilePresetsCollection) GetName() string {
	return "profile_presets"
}

// GetRoute is where the collection is in its URL, /presets/profiles
func (s *ProfilePresetsCollection) GetRoute() string {
	return "profiles"
}

// UpdateRow renames a profile, keeping each name in its profile store different, as Nightscout keys them by name
func (s *ProfilePresetsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteProfilePreset, model *sitemodels.SiteProfilePreset) (*sitedb.ProfilePreset, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	if model.Name != nil {
		if *model.Name == "" {
			return nil, apiutils.ErrMissingRequiredFieldWithHint("name")
		}
		store, err := q.GetProfileStore(ctx, sitedb.GetProfileStoreParams{ProfileStoreID: oldModel.ProfileStoreID})
		if err != nil {
			return nil, err
		}
		for _, profile := range store {
			if profile.ID != oldModel.ID && profile.Name == *model.Name {
				return nil, apiutils.ErrConstraintCheckFailedWithHint("another profile in this store is called " + *model.Name)
			}
		}
	}
	model.ID = oldModel.ID
	return q.UpdateProfilePreset(ctx, sitemodels.ProfilePresetToUpdate(model, siteContext))
}

func (s *ProfilePresetsCollection) UpdateHTMXResponse(model *sitemodels.SiteProfilePreset) templ.Component {
	return site_presets_templ.ProfilePresetUpdated(model)
}

func (s *ProfilePresetsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.ProfilePreset, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetProfilePreset, sitedb.GetProfilePresetParams{ID: id})
}

// ToModel is the site model of a row, its sensitivities and targets in the settings' units
func (s *ProfilePresetsCollection) ToModel(row *sitedb.ProfilePreset, siteContext *httputils.SiteContext) *sitemodels.SiteProfilePreset {
	return sitemodels.ProfilePresetFromDB(row, siteContext)
}

// GetHTMXResponse is the profile's details and schedules, shown in a modal
func (s *ProfilePresetsCollection) GetHTMXResponse(model *sitemodels.SiteProfilePreset) templ.Component {
	return site_presets_templ.ProfilePresetModal(model)
}

func (s *ProfilePresetsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.ProfilePreset, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllProfilePresets, siteapiutils.QueryToAllProfilePresets(query))
}

func (s *ProfilePresetsCollection) SearchHTMXResponse(models []*sitemodels.SiteProfilePreset, query *apiutils.Query) templ.Component {
	return site_presets_templ.ProfilePresetsTable(models, query.Skip, query.Limit)
}

// DeleteRow hard deletes the one profile, as the Nightscout API does presets, leaving the rest of its store
func (s *ProfilePresetsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	_, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).HardDeleteProfilePreset, sitedb.HardDeleteProfilePresetParams{ID: id})
	return err
}

func (s *ProfilePresetsCollection) DeleteHTMXResponse(model *sitemodels.SiteProfilePreset) templ.Component {
	return site_presets_templ.ProfilePresetDeleted()
}
