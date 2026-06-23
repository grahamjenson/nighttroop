package nightscout_api

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"

	"nighttroop.t1d.nz/utils/config"
)

type CollectionStore interface {
	Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError)
	Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError)
	Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError)
	Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError)
	Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError
	DeleteType(ctx context.Context, typeString string, permanent bool) (int, *nsutils.HTTPError)
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
	record, err := collectionsStore.Create(ctx, data)

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
	record, err := collectionsStore.Get(ctx, identifier)
	if err != nil {
		return nsutils.Record{}, err
	}
	if record == nil {
		return nsutils.Record{}, nsutils.ErrNotFound
	}
	return *record, nil
}

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection nsutils.Collection, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	query = withDefaultQuery(query)

	if !ok {
		return nil, nsutils.ErrInvalidCollection
	}
	records, err := collectionsStore.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	records = nsutils.ApplyQuery(records, query)
	return records, nil

}

func (nts *NightTroopAPIStore) Update(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any) (nsutils.Record, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, nsutils.ErrInvalidCollection
	}
	record, err := collectionsStore.Update(ctx, identifier, patch)
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
	return collectionsStore.Delete(ctx, identifier, permanent)
}

func (nts *NightTroopAPIStore) DeleteType(ctx context.Context, collection nsutils.Collection, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return 0, nsutils.ErrInvalidCollection
	}
	return collectionsStore.DeleteType(ctx, typeString, permanent)

}

func withDefaultQuery(query *nsutils.Query) *nsutils.Query {
	defaults := nsutils.DefaultQuery()
	if query.SortField == "" {
		query.SortField = defaults.SortField
	}
	if query.Limit == 0 {
		query.Limit = defaults.Limit
	}
	return query
}
