package nightscoutapi

import (
	"cmp"
	"context"
	"log/slog"
	"maps"
	"reflect"
	"slices"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nscollections"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

type NightTroopAPIStore struct {
	controlDB *db.ControlDB
	config    *config.ServerConfig

	stores map[nsutils.Collection]nscollections.CollectionStore
}

func NewNightTroopAPIStore(controlDB *db.ControlDB, config *config.ServerConfig) *NightTroopAPIStore {
	return &NightTroopAPIStore{
		controlDB: controlDB,
		config:    config,
		stores:    AllCollections(),
	}
}

func AllCollections() map[nsutils.Collection]nscollections.CollectionStore {
	return map[nsutils.Collection]nscollections.CollectionStore{
		nsutils.Entries:      nscollections.NewEntriesCollection(),
		nsutils.Treatments:   nscollections.NewTreatmentCollection(),
		nsutils.Devicestatus: nscollections.NewDeviceStatusCollection(),
		nsutils.Profile:      nscollections.NewProfileCollection(),
		nsutils.Food:         nscollections.NewFoodCollection(),
		nsutils.Settings:     nscollections.NewSettingsCollection(),
		nsutils.Activity:     nscollections.NewActivityCollection(),
	}
}

func (nts *NightTroopAPIStore) GetStore(collection nsutils.Collection) (nscollections.CollectionStore, bool) {
	collectionsStore, ok := nts.stores[collection]
	return collectionsStore, ok
}

// collectionFor is the collection's store and the request's site, or the error when either is missing
func (nts *NightTroopAPIStore) collectionFor(ctx context.Context, collection nsutils.Collection) (nscollections.CollectionStore, *httputils.SiteContext, *nsutils.HTTPError) {
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

// found is the document, or not found when there is none
func found(doc nsutils.Document, err *nsutils.HTTPError) (nsutils.Document, *nsutils.HTTPError) {
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nsutils.ErrNotFound
	}
	return doc, nil
}

// foundAnd is found with the flag (created or duplicated) that came with the record
func foundAnd(doc nsutils.Document, flag bool, err *nsutils.HTTPError) (nsutils.Document, bool, *nsutils.HTTPError) {
	r, err := found(doc, err)
	return r, flag && err == nil, err
}

// Return Model that has identifier and srv modified and also a boolean indicating if the record was created (true) or updated (false)
func (nts *NightTroopAPIStore) Create(ctx context.Context, collection nsutils.Collection, data map[string]any) (nsutils.Document, bool, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, false, err
	}
	learnUnitsLabels(ctx, siteContext, data)
	return foundAnd(collectionsStore.Create(ctx, siteContext, data))
}

// unitsSpellings are how apps spell both units, by either: as Nightscout does, mg/dl and mmol, as AAPS, Trio and
// xDrip4iOS do, or as Nighttroop does, mg/dL and mmol/L, as Loop does
var unitsSpellings = map[string]struct{ mgdl, mmol string }{
	"mg/dl": {"mg/dl", "mmol"}, "mmol": {"mg/dl", "mmol"},
	"mg/dL": {"mg/dL", "mmol/L"}, "mmol/L": {"mg/dL", "mmol/L"},
}

// learnUnitsLabels gives a device units back spelt as it first sends them, unless its spelling is set, as apps read
// only their own: xDrip4iOS reads a BG check's glucose as mg/dL only when its units are exactly mg/dl, so it saw 110 come
// back as 1982 and uploaded it again every 15 seconds; AAPS knows only mg/dl and mmol; Loop reads only mg/dL
func learnUnitsLabels(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) {
	sent, _ := data["units"].(string)
	spelling, ok := unitsSpellings[sent]
	identity := siteContext.IdentityModel
	if !ok || identity == nil || (identity.UnitsMgdlLabel != nil && identity.UnitsMmolLabel != nil) {
		return
	}
	learned := *identity
	learned.UnitsMgdlLabel = cmp.Or(learned.UnitsMgdlLabel, &spelling.mgdl)
	learned.UnitsMmolLabel = cmp.Or(learned.UnitsMmolLabel, &spelling.mmol)
	err := siteContext.ControlDB.Queries.LearnSiteIdentityUnitsLabels(ctx, controldb.LearnSiteIdentityUnitsLabelsParams{
		SiteID: siteContext.SiteID, ID: learned.ID, UnitsMgdlLabel: learned.UnitsMgdlLabel, UnitsMmolLabel: learned.UnitsMmolLabel})
	if err != nil {
		slog.ErrorContext(ctx, "failed to keep the identity's units labels", "error", err)
		return
	}
	httputils.SetIdentityModelInSiteContext(ctx, &learned)
}

func (nts *NightTroopAPIStore) Get(ctx context.Context, collection nsutils.Collection, identifier string, query nsquery.Query) (nsutils.Document, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, err
	}
	doc, httpErr := found(collectionsStore.Get(ctx, siteContext, identifier, query))
	// A treatment's carbs part, which API v3 gives as its own document, is found by its identifier
	if treatment, isPart := nscollections.CarbsPartOf(identifier); httpErr == nsutils.ErrNotFound && isPart && collection == nsutils.Treatments {
		if parts := nsutils.V3Parts(foundParts(nts.Get(ctx, collection, treatment, query))); len(parts) == 2 {
			return parts[1], nil
		}
	}
	if reading, ok := nts.fingerStick(ctx, siteContext, collection, identifier); httpErr == nsutils.ErrNotFound && ok {
		if bgCheck, ok := nts.bgCheck(reading, siteContext); ok {
			return bgCheck, nil
		}
	}
	return doc, httpErr
}

// foundParts is a document found as a list, empty when none was
func foundParts(doc nsutils.Document, err *nsutils.HTTPError) []nsutils.Document {
	if err != nil {
		return nil
	}
	return []nsutils.Document{doc}
}

// SplitV3 makes a treatment API v3 gives as its insulin and its carbs two treatments, before a change or delete of
// either through API v3, so it changes only that one, as AAPS changes and deletes them on their own: its carbs become a
// treatment with the carbs part's identifier, and are taken out of it
func (nts *NightTroopAPIStore) SplitV3(ctx context.Context, collection nsutils.Collection, identifier string) *nsutils.HTTPError {
	treatment, isPart := nscollections.CarbsPartOf(identifier)
	if !isPart {
		treatment = identifier
	}
	if collection != nsutils.Treatments {
		return nil
	}
	parts := nsutils.V3Parts(foundParts(nts.Get(ctx, collection, treatment, nsquery.Query{})))
	if len(parts) < 2 {
		return nil
	}
	carbs := parts[1].Fields(true)
	for _, server := range []string{"srvCreated", "srvModified", "subject", "isValid", "isReadOnly"} {
		delete(carbs, server)
	}
	if _, _, err := nts.Create(ctx, collection, carbs); err != nil {
		return err
	}
	_, err := nts.Patch(ctx, collection, treatment, map[string]any{"carbs": nil, "protein": nil, "fat": nil, "foodType": nil, "absorptionTime": nil})
	return err
}

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection nsutils.Collection, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, err
	}
	if query.UnknownFilter {
		return []nsutils.Document{}, nil
	}
	if collection != nsutils.Treatments {
		return collectionsStore.Search(ctx, siteContext, query)
	}
	return nts.withFingerSticks(ctx, siteContext, query, func(query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
		docs, err := collectionsStore.Search(ctx, siteContext, query)
		return withoutCarbsParts(ctx, collectionsStore, siteContext, query, docs, err)
	})
}

func (nts *NightTroopAPIStore) History(ctx context.Context, collection nsutils.Collection, since int64, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, err
	}
	if query.UnknownFilter {
		return []nsutils.Document{}, nil
	}
	docs, httpErr := collectionsStore.History(ctx, siteContext, since, query)
	if httpErr != nil || collection != nsutils.Treatments {
		return docs, httpErr
	}
	docs, _ = withoutCarbsParts(ctx, collectionsStore, siteContext, query, docs, nil)
	return nts.historyWithFingerSticks(ctx, siteContext, since, query, docs)
}

// withoutCarbsParts are the treatments found, each one whose carbs part is stored as its own treatment, see SplitV3,
// as its insulin part only on API v3, so a treatment that got its carbs again does not give two with that identifier
func withoutCarbsParts(ctx context.Context, store nscollections.CollectionStore, siteContext *httputils.SiteContext, query nsquery.Query, docs []nsutils.Document, err *nsutils.HTTPError) ([]nsutils.Document, *nsutils.HTTPError) {
	if err != nil || !query.V3 {
		return docs, err
	}
	for i, doc := range docs {
		parts := nsutils.V3Parts([]nsutils.Document{doc})
		if len(parts) < 2 {
			continue
		}
		if _, err := store.Get(ctx, siteContext, parts[1].Identifier(), nsquery.Query{IncludeDeleted: true}); err == nil {
			docs[i] = parts[0]
		}
	}
	return docs, nil
}

func (nts *NightTroopAPIStore) Replace(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any, canCreate bool) (nsutils.Document, bool, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, false, err
	}
	if collection == nsutils.Treatments {
		if patch, err = nts.withoutCarbsPart(ctx, collectionsStore, siteContext, identifier, patch); err != nil {
			return nil, false, err
		}
	}
	return foundAnd(collectionsStore.Replace(ctx, siteContext, identifier, patch, canCreate))
}

// carbsFields are the fields of a treatment its carbs part has, see SplitV3
var carbsFields = []string{"carbs", "protein", "fat", "foodType", "absorptionTime"}

// withoutCarbsPart is a treatment sent with carbs without them, once its carbs part is stored as its own treatment, see
// SplitV3, as xDrip+ sends a treatment again when it is edited or its upload queue sends it again. Its carbs go to the
// carbs part, so they are not counted twice. If the carbs part is deleted, it stays deleted whatever carbs are sent, as
// AAPS deleted it and the loop would otherwise dose for them. The treatment keeps the event type it has with its carbs.
func (nts *NightTroopAPIStore) withoutCarbsPart(ctx context.Context, store nscollections.CollectionStore, siteContext *httputils.SiteContext, identifier string, data map[string]any) (map[string]any, *nsutils.HTTPError) {
	partID := nscollections.CarbsPart(identifier)
	var whole nightscoutmodels.Treatment
	if data["carbs"] == nil || utils.ConvertViaJSON(data, &whole) != nil || whole.EventType == nil || whole.Carbs == nil {
		return data, nil
	}
	part, err := store.Get(ctx, siteContext, partID, nsquery.Query{IncludeDeleted: true})
	if err == nsutils.ErrNotFound {
		return data, nil
	} else if err != nil {
		return nil, err
	}
	carbs, rest, stored := map[string]any{}, maps.Clone(data), part.Fields(true)
	for _, name := range carbsFields {
		carbs[name] = data[name]
		delete(rest, name)
	}
	rest["eventType"] = whole.EventType.String()
	if stored["isValid"] == true && slices.ContainsFunc(carbsFields, func(name string) bool { return !reflect.DeepEqual(stored[name], carbs[name]) }) {
		_, err = nts.Patch(ctx, nsutils.Treatments, partID, carbs)
	}
	return rest, err
}

func (nts *NightTroopAPIStore) Patch(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any) (nsutils.Document, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return nil, err
	}
	return found(collectionsStore.Patch(ctx, siteContext, identifier, patch))
}

func (nts *NightTroopAPIStore) Delete(ctx context.Context, collection nsutils.Collection, identifier string, permanent bool) *nsutils.HTTPError {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return err
	}
	httpErr := collectionsStore.Delete(ctx, siteContext, identifier, permanent)
	// A BG Check given for a meter reading is deleted with its reading, as a finger stick deleted in AAPS
	if reading, ok := nts.fingerStick(ctx, siteContext, collection, identifier); httpErr == nsutils.ErrNotFound && ok {
		return nts.Delete(ctx, nsutils.Entries, reading.ID, permanent)
	}
	return httpErr
}

func (nts *NightTroopAPIStore) DeleteType(ctx context.Context, collection nsutils.Collection, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return 0, err
	}
	return collectionsStore.DeleteType(ctx, siteContext, typeString, permanent)
}

func (nts *NightTroopAPIStore) DeleteQuery(ctx context.Context, collection nsutils.Collection, query nsquery.Query) (int, *nsutils.HTTPError) {
	collectionsStore, siteContext, err := nts.collectionFor(ctx, collection)
	if err != nil {
		return 0, err
	}
	return collectionsStore.DeleteQuery(ctx, siteContext, query)
}
