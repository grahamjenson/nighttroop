package sitecollections

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
)

type SiteCollectionRoutes struct {
	Search http.HandlerFunc
	Get    http.HandlerFunc
	Create http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
}

type SiteStore struct {
	collectionRoutes map[string]*SiteCollectionRoutes
}

func NewSiteStore() *SiteStore {

	siteStore := &SiteStore{

		collectionRoutes: make(map[string]*SiteCollectionRoutes),
	}

	// Data collections
	siteStore.registerCollection(&sitedata.SummaryCollection{})
	siteStore.registerCollection(&sitedata.GlucoseReadingsCollection{})
	siteStore.registerCollection(&sitedata.TempTargetsCollection{})
	siteStore.registerCollection(&sitedata.TempBasalsCollection{})
	siteStore.registerCollection(&sitedata.ActivitiesCollection{})
	siteStore.registerCollection(&sitedata.ComponentEventsCollection{})
	siteStore.registerCollection(&sitedata.NotificationEventsCollection{})
	siteStore.registerCollection(&sitedata.ProfileSwitchesCollection{})
	siteStore.registerCollection(&sitedata.EventGroupsCollection{})
	siteStore.registerCollection(&sitedata.InsulinDosesCollection{})
	siteStore.registerCollection(&sitedata.FoodIntakesCollection{})
	siteStore.registerCollection(&sitedata.DeviceStatusesCollection{})

	// Settings collections
	siteStore.registerCollection(&sitesettings.SyncSettingsCollection{})
	siteStore.registerCollection(&sitesettings.IdentitiesCollection{})
	siteStore.registerCollection(&sitesettings.SettingsCollection{Site: true})
	siteStore.registerCollection(&sitesettings.SettingsCollection{Site: false})

	return siteStore
}

func (store *SiteStore) registerCollection[R, M any](collection CollectionBase[R, M]) {
	name := collection.GetName()
	route := collection.GetRoute()
	if _, exists := store.collectionRoutes[route]; exists {
		panic("collection route already registered: " + route)
	}

	routes := &SiteCollectionRoutes{
		Search: notFoundHandler,
		Get:    notFoundHandler,
		Create: notFoundHandler,
		Update: notFoundHandler,
		Delete: notFoundHandler,
	}

	added := false
	if val, ok := collection.(CollectionSearch[R, M]); ok {
		routes.Search = SearchHandler(val)
		added = true
	}
	if val, ok := collection.(CollectionGet[R, M]); ok {
		routes.Get = GetHandler(val)
		added = true
	}
	if val, ok := collection.(CollectionCreate[R, M]); ok {
		routes.Create = CreateHandler(val)
		added = true
	}
	if val, ok := collection.(CollectionUpdate[R, M]); ok {
		routes.Update = UpdateHandler(val)
		added = true
	}
	if val, ok := collection.(CollectionUpsert[R, M]); ok {
		routes.Update = UpsertHandler(val)
	}
	if val, ok := collection.(CollectionDelete[R, M]); ok {
		routes.Delete = DeleteHandler(val)
		added = true
	}

	if !added {
		panic("collection does not implement any known interfaces: " + name)
	}

	store.collectionRoutes[route] = routes

}

// Collection is the routes of the collection at route, for a route of its own like the summary at /
func (s *SiteStore) Collection(route string) *SiteCollectionRoutes {
	routes, exists := s.collectionRoutes[route]
	if !exists {
		panic("no collection at route: " + route)
	}
	return routes
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "collection method not found", http.StatusNotFound)
}

func notFoundHandlerCollection(w http.ResponseWriter, r *http.Request, name string) {
	http.Error(w, "collection method not found: "+name, http.StatusNotFound)
}

func (s *SiteStore) CollectionSearchHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Search(w, r)
		return
	}
	notFoundHandlerCollection(w, r, name)
}
func (s *SiteStore) CollectionGetHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Get(w, r)
		return
	}
	notFoundHandlerCollection(w, r, name)
}
func (s *SiteStore) CollectionCreateHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Create(w, r)
		return
	}
	notFoundHandlerCollection(w, r, name)
}
func (s *SiteStore) CollectionUpdateHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Update(w, r)
		return
	}
	notFoundHandlerCollection(w, r, name)
}
func (s *SiteStore) CollectionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Delete(w, r)
		return
	}
	notFoundHandlerCollection(w, r, name)
}
