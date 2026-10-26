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
	collectionRoutes map[string]SiteCollectionRoutes
}

func NewSiteStore() *SiteStore {

	siteStore := &SiteStore{

		collectionRoutes: make(map[string]SiteCollectionRoutes),
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
	if _, exists := store.collectionRoutes[name]; exists {
		panic("collection already registered: " + name)
	}

	routes := SiteCollectionRoutes{
		Search: notFoundHandler,
		Get:    notFoundHandler,
		Create: notFoundHandler,
		Update: notFoundHandler,
		Delete: notFoundHandler,
	}

	if val, ok := collection.(CollectionSearch[R, M]); ok {
		routes.Search = SearchHandler(val)
	}
	if val, ok := collection.(CollectionGet[R, M]); ok {
		routes.Get = GetHandler(val)
	}
	if val, ok := collection.(CollectionCreate[R, M]); ok {
		routes.Create = CreateHandler(val)
	}
	if val, ok := collection.(CollectionUpdate[R, M]); ok {
		routes.Update = UpdateHandler(val)
	}
	if val, ok := collection.(CollectionDelete[R, M]); ok {
		routes.Delete = DeleteHandler(val)
	}
	store.collectionRoutes[name] = routes

}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "collection method not found", http.StatusNotFound)
}

func (s *SiteStore) CollectionSearchHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Search(w, r)
		return
	}
	notFoundHandler(w, r)
}
func (s *SiteStore) CollectionGetHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Get(w, r)
		return
	}
	notFoundHandler(w, r)
}
func (s *SiteStore) CollectionCreateHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Create(w, r)
		return
	}
	notFoundHandler(w, r)
}
func (s *SiteStore) CollectionUpdateHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Update(w, r)
		return
	}
	notFoundHandler(w, r)
}
func (s *SiteStore) CollectionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[name]; exists {
		routes.Delete(w, r)
		return
	}
	notFoundHandler(w, r)
}
