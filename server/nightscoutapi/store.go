package nightscoutapi

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscoutapi/collections"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

type NightTroopAPIStore struct {
	controlDB *db.ControlDB
	config    *config.ServerConfig

	stores map[nsutils.Collection]collections.CollectionStore

	// onChange is told of each record created or changed, for live updates
	onChange func(ctx context.Context, collection nsutils.Collection, record nsutils.Record)
}

func NewNightTroopAPIStore(controlDB *db.ControlDB, config *config.ServerConfig) *NightTroopAPIStore {
	return &NightTroopAPIStore{
		controlDB: controlDB,
		config:    config,
		stores: map[nsutils.Collection]collections.CollectionStore{
			nsutils.Entries:      collections.NewEntriesCollection(),
			nsutils.Treatments:   collections.NewTreatmentCollection(),
			nsutils.Devicestatus: collections.NewDeviceStatusCollection(),
			nsutils.Profile:      collections.NewProfileCollection(),
			nsutils.Food:         collections.NewFoodCollection(),
			nsutils.Settings:     collections.NewSettingsCollection(),
			nsutils.Activity:     collections.NewActivityCollection(),
			// Add more collections as needed
		},
	}
}

func (nts *NightTroopAPIStore) GetStore(collection nsutils.Collection) (collections.CollectionStore, bool) {
	collectionsStore, ok := nts.stores[collection]
	return collectionsStore, ok
}

// collectionFor is the collection's store and the request's site, or the error when either is missing
func (nts *NightTroopAPIStore) collectionFor(ctx context.Context, collection nsutils.Collection) (collections.CollectionStore, *httputils.SiteContext, *nsutils.HTTPError) {
	collectionsStore, ok := nts.GetStore(collection)
	if !ok {
		return nil, nil, nsutils.ErrInvalidCollection
	}
	siteContext := httputils.GetSiteContext(ctx)
	if siteContext == nil {
		return nil, nil, nsutils.ErrMissingCredentials
	}
	return collectionsStore, siteContext, nil
}

// found is the record, or not found when there is none
func found(record *nsutils.Record, err *nsutils.HTTPError) (nsutils.Record, *nsutils.HTTPError) {
	if err != nil {
		return nsutils.Record{}, err
	}
	if record == nil {
		return nsutils.Record{}, nsutils.ErrNotFound
	}
	return *record, nil
}

// foundAnd is found with the flag (created or duplicated) that came with the record
func foundAnd(record *nsutils.Record, flag bool, err *nsutils.HTTPError) (nsutils.Record, bool, *nsutils.HTTPError) {
	r, err := found(record, err)
	return r, flag && err == nil, err
}

// Return Model that has identifier and srv modified and also a boolean indicating if the record was created (true) or updated (false)
func (nts *NightTroopAPIStore) Create(ctx context.Context, collection nsutils.Collection, data map[string]any, raw bool) (nsutils.Record, bool, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nsutils.Record{}, false, err
	}
	record, duplicate, err := foundAnd(collectionsStore.Create(ctx, siteContext, data, raw))
	if err == nil && !duplicate {
		nts.changed(ctx, collection, record)
	}
	return record, duplicate, err
}

func (nts *NightTroopAPIStore) changed(ctx context.Context, collection nsutils.Collection, record nsutils.Record) {
	if nts.onChange != nil {
		nts.onChange(ctx, collection, record)
	}
}

func (nts *NightTroopAPIStore) Get(ctx context.Context, collection nsutils.Collection, identifier string, query nsquery.Query) (nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nsutils.Record{}, err
	}
	return found(collectionsStore.Get(ctx, siteContext, identifier, query))
}

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection nsutils.Collection, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, err
	}
	return collectionsStore.Search(ctx, siteContext, query)
}

func (nts *NightTroopAPIStore) Replace(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any, canCreate bool, raw bool) (nsutils.Record, bool, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nsutils.Record{}, false, err
	}
	record, created, err := foundAnd(collectionsStore.Replace(ctx, siteContext, identifier, patch, canCreate, raw))
	if err == nil {
		nts.changed(ctx, collection, record)
	}
	return record, created, err
}

func (nts *NightTroopAPIStore) Patch(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any, raw bool) (nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nsutils.Record{}, err
	}
	record, err := found(collectionsStore.Patch(ctx, siteContext, identifier, patch, raw))
	if err == nil {
		nts.changed(ctx, collection, record)
	}
	return record, err
}

func (nts *NightTroopAPIStore) Delete(ctx context.Context, collection nsutils.Collection, identifier string, permanent bool) *nsutils.HTTPError {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return err
	}
	return collectionsStore.Delete(ctx, siteContext, identifier, permanent)
}

func (nts *NightTroopAPIStore) DeleteType(ctx context.Context, collection nsutils.Collection, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return 0, err
	}
	return collectionsStore.DeleteType(ctx, siteContext, typeString, permanent)
}
