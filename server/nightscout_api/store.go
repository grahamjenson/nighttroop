package nightscout_api

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

type CollectionStore interface {
	Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError)
	Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError)
	Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError)
	Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError)
	Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError
	DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError)
}

type NightTroopAPIStore struct {
	controlDB *db.ControlDB
	config    config.ServerConfig

	stores map[nsutils.Collection]CollectionStore
}

// Return Model that has identifier and srv modified and also a boolean indicating if the record was created (true) or updated (false)
func (nts *NightTroopAPIStore) Create(ctx context.Context, collection nsutils.Collection, data map[string]any) (nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return nsutils.Record{}, nsutils.ErrMissingCredentials
	}
	record, err := collectionsStore.Create(ctx, siteContext, data)

	if err != nil {
		return nsutils.Record{}, err
	}
	if record == nil {
		return nsutils.Record{}, nsutils.ErrNotFound
	}
	return *record, nil

}

func (nts *NightTroopAPIStore) Get(ctx context.Context, collection nsutils.Collection, identifier string) (nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return nsutils.Record{}, nsutils.ErrMissingCredentials
	}
	record, err := collectionsStore.Get(ctx, siteContext, identifier)
	if err != nil {
		return nsutils.Record{}, err
	}
	if record == nil {
		return nsutils.Record{}, nsutils.ErrNotFound
	}
	return *record, nil
}

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection nsutils.Collection, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]

	if !ok {
		return nil, nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return nil, nsutils.ErrMissingCredentials
	}
	records, err := collectionsStore.Search(ctx, siteContext, query)
	if err != nil {
		return nil, err
	}

	return records, nil

}

func (nts *NightTroopAPIStore) Update(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any) (nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return nsutils.Record{}, nsutils.ErrMissingCredentials
	}
	record, err := collectionsStore.Update(ctx, siteContext, identifier, patch)
	if err != nil {
		return nsutils.Record{}, err
	}
	if record == nil {
		return nsutils.Record{}, nsutils.ErrNotFound
	}
	return *record, nil
}

func (nts *NightTroopAPIStore) Delete(ctx context.Context, collection nsutils.Collection, identifier string, permanent bool) *nsutils.HTTPError {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return nsutils.ErrMissingCredentials
	}
	return collectionsStore.Delete(ctx, siteContext, identifier, permanent)
}

func (nts *NightTroopAPIStore) DeleteType(ctx context.Context, collection nsutils.Collection, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return 0, nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return 0, nsutils.ErrMissingCredentials
	}
	return collectionsStore.DeleteType(ctx, siteContext, typeString, permanent)

}
