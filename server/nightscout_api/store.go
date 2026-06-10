package nightscout_api

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
	"nighttroop.t1d.nz/utils/config"
)

type NSCollection string

const (
	EntriesCollection      NSCollection = "entries"
	TreatmentsCollection   NSCollection = "treatments"
	ProfileCollection      NSCollection = "profile"
	DeviceStatusCollection NSCollection = "devicestatus"
	SettingsCollection     NSCollection = "settings"
	FoodCollection         NSCollection = "food"
	PropertiesCollection   NSCollection = "properties"
)

type CollectionStore interface {
	Create(ctx context.Context, data map[string]any) (*model.Record, bool, error)
	Get(ctx context.Context, identifier string) (*model.Record, error)
	Search(ctx context.Context, query di.Query) ([]model.Record, error)
	Update(ctx context.Context, collection, identifier string, patch map[string]any) (*model.Record, error)
	Delete(ctx context.Context, collection, identifier string, permanent bool) error
	DeleteMatching(ctx context.Context, collection string, query di.Query, permanent bool) (int, error)
}

type NightTroopAPIStore struct {
	controlDB *db.ControlDB
	config    config.ServerConfig

	stores map[NSCollection]CollectionStore
}

// Return Model that has identifier and srv modified and also a boolean indicating if the record was created (true) or updated (false)
func (nts *NightTroopAPIStore) Create(ctx context.Context, collection string, data map[string]any, subject string) (model.Record, bool, error) {
	nsCollection := normalizeCollection(collection)
	collectionsStore, ok := nts.stores[nsCollection]
	if !ok {
		return model.Record{}, false, fmt.Errorf("unknown collection: %s", collection)
	}
	record, created, err := collectionsStore.Create(ctx, data)

	if err != nil {
		return model.Record{}, false, err
	}
	if record == nil {
		return model.Record{}, false, di.ErrNotFound
	}
	return *record, created, nil

}

func (nts *NightTroopAPIStore) Get(ctx context.Context, collection, identifier string) (model.Record, error) {
	nsCollection := normalizeCollection(collection)
	collectionsStore, ok := nts.stores[nsCollection]
	if !ok {
		return model.Record{}, fmt.Errorf("unknown collection: %s", collection)
	}
	record, err := collectionsStore.Get(ctx, identifier)
	if err != nil {
		return model.Record{}, err
	}
	if record == nil {
		return model.Record{}, di.ErrNotFound
	}
	return *record, nil
}

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection string, query di.Query) ([]model.Record, error) {
	nsCollection := normalizeCollection(collection)
	query = withDefaultQuery(query)

	collectionsStore, ok := nts.stores[nsCollection]
	if !ok {
		return nil, fmt.Errorf("unknown collection: %s", collection)
	}
	records, err := collectionsStore.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	records = di.ApplyQuery(records, query)
	return records, nil

}

func (nts *NightTroopAPIStore) Update(ctx context.Context, collection, identifier string, patch map[string]any, subject string) (model.Record, error) {
	nsCollection := normalizeCollection(collection)
	collectionsStore, ok := nts.stores[nsCollection]
	if !ok {
		return model.Record{}, fmt.Errorf("unknown collection: %s", collection)
	}
	record, err := collectionsStore.Update(ctx, collection, identifier, patch)
	if err != nil {
		return model.Record{}, err
	}
	if record == nil {
		return model.Record{}, di.ErrNotFound
	}
	return *record, nil
}

func (nts *NightTroopAPIStore) Delete(ctx context.Context, collection, identifier string, permanent bool, subject string) error {
	nsCollection := normalizeCollection(collection)
	collectionsStore, ok := nts.stores[nsCollection]
	if !ok {
		return fmt.Errorf("unknown collection: %s", collection)
	}
	return collectionsStore.Delete(ctx, collection, identifier, permanent)
}

func (nts *NightTroopAPIStore) DeleteMatching(ctx context.Context, collection string, query di.Query, permanent bool, subject string) (int, error) {
	nsCollection := normalizeCollection(collection)
	collectionsStore, ok := nts.stores[nsCollection]
	if !ok {
		return 0, fmt.Errorf("unknown collection: %s", collection)
	}
	return collectionsStore.DeleteMatching(ctx, collection, query, permanent)
}

func normalizeCollection(collection string) NSCollection {
	collection = model.NormalizeCollection(collection)
	return NSCollection(collection)
}

func withDefaultQuery(query di.Query) di.Query {
	defaults := di.DefaultQuery()
	if query.SortField == "" {
		query.SortField = defaults.SortField
	}
	if query.Limit == 0 {
		query.Limit = defaults.Limit
	}
	return query
}
