package nightscout_api

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"

	"nighttroop.t1d.nz/utils/config"
)

type CollectionStore interface {
	Create(ctx context.Context, data map[string]any) (*nsutils.Record, bool, error)
	Get(ctx context.Context, identifier string) (*nsutils.Record, error)
	Search(ctx context.Context, query nsutils.Query) ([]nsutils.Record, error)
	Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, error)
	Delete(ctx context.Context, identifier string, permanent bool) error
	DeleteMatching(ctx context.Context, query nsutils.Query, permanent bool) (int, error)
}

type NightTroopAPIStore struct {
	controlDB *db.ControlDB
	config    config.ServerConfig

	stores map[nsutils.Collection]CollectionStore
}

// Return Model that has identifier and srv modified and also a boolean indicating if the record was created (true) or updated (false)
func (nts *NightTroopAPIStore) Create(ctx context.Context, collection nsutils.Collection, data map[string]any, subject string) (nsutils.Record, bool, error) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, false, fmt.Errorf("unknown collection: %s", collection)
	}
	record, created, err := collectionsStore.Create(ctx, data)

	if err != nil {
		return nsutils.Record{}, false, err
	}
	if record == nil {
		return nsutils.Record{}, false, nsutils.ErrNotFound
	}
	return *record, created, nil

}

func (nts *NightTroopAPIStore) Get(ctx context.Context, collection nsutils.Collection, identifier string) (nsutils.Record, error) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, fmt.Errorf("unknown collection: %s", collection)
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

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection nsutils.Collection, query nsutils.Query) ([]nsutils.Record, error) {
	collectionsStore, ok := nts.stores[collection]
	query = withDefaultQuery(query)

	if !ok {
		return nil, fmt.Errorf("unknown collection: %s", collection)
	}
	records, err := collectionsStore.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	records = nsutils.ApplyQuery(records, query)
	return records, nil

}

func (nts *NightTroopAPIStore) Update(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any, subject string) (nsutils.Record, error) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return nsutils.Record{}, fmt.Errorf("unknown collection: %s", collection)
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

func (nts *NightTroopAPIStore) Delete(ctx context.Context, collection nsutils.Collection, identifier string, permanent bool, subject string) error {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return fmt.Errorf("unknown collection: %s", collection)
	}
	return collectionsStore.Delete(ctx, identifier, permanent)
}

func (nts *NightTroopAPIStore) DeleteMatching(ctx context.Context, collection nsutils.Collection, query nsutils.Query, permanent bool, subject string) (int, error) {
	collectionsStore, ok := nts.stores[collection]
	if !ok {
		return 0, fmt.Errorf("unknown collection: %s", collection)
	}
	return collectionsStore.DeleteMatching(ctx, query, permanent)
}

func withDefaultQuery(query nsutils.Query) nsutils.Query {
	defaults := nsutils.DefaultQuery()
	if query.SortField == "" {
		query.SortField = defaults.SortField
	}
	if query.Limit == 0 {
		query.Limit = defaults.Limit
	}
	return query
}
