package sitecollections

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitepresets"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type SiteCollectionRoutes struct {
	Search http.HandlerFunc
	Get    http.HandlerFunc
	Create http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
}

type SiteStore struct {
	// collectionRoutes are each section's collections by route, e.g. data then glucose for /data/glucose
	collectionRoutes map[string]map[string]*SiteCollectionRoutes
}

func NewSiteStore() *SiteStore {

	siteStore := &SiteStore{

		collectionRoutes: make(map[string]map[string]*SiteCollectionRoutes),
	}

	// Data collections
	siteStore.registerCollection("data", &sitedata.SummaryCollection{})
	siteStore.registerCollection("data", &sitedata.GlucoseReadingsCollection{})
	siteStore.registerCollection("data", &sitedata.TempTargetsCollection{})
	siteStore.registerCollection("data", &sitedata.TempBasalsCollection{})
	siteStore.registerCollection("data", &sitedata.ActivitiesCollection{})
	siteStore.registerCollection("data", &sitedata.ComponentEventsCollection{})
	siteStore.registerCollection("data", &sitedata.NotificationEventsCollection{})
	siteStore.registerCollection("data", &sitedata.ProfileSwitchesCollection{})
	siteStore.registerCollection("data", &sitedata.EventGroupsCollection{})
	siteStore.registerCollection("data", &sitedata.InsulinDosesCollection{})
	siteStore.registerCollection("data", &sitedata.FoodIntakesCollection{})
	siteStore.registerCollection("data", &sitedata.DeviceStatusesCollection{})

	// Presets collections
	siteStore.registerCollection("presets", &sitepresets.FoodPresetsCollection{})
	siteStore.registerCollection("presets", &sitepresets.ProfilePresetsCollection{})

	// Settings collections
	siteStore.registerCollection("settings", &sitesettings.SyncSettingsCollection{})
	siteStore.registerCollection("settings", &sitesettings.IdentitiesCollection{})
	siteStore.registerCollection("settings", &sitesettings.SettingsCollection{Site: true})
	siteStore.registerCollection("settings", &sitesettings.SettingsCollection{Site: false})

	return siteStore
}

func (store *SiteStore) registerCollection[R, M any](section string, collection CollectionBase[R, M]) {
	name := collection.GetName()
	route := collection.GetRoute()
	if store.collectionRoutes[section] == nil {
		store.collectionRoutes[section] = make(map[string]*SiteCollectionRoutes)
	}
	if _, exists := store.collectionRoutes[section][route]; exists {
		panic("collection route already registered: " + section + "/" + route)
	}

	cannot := func(what string) http.HandlerFunc {
		return answerNotFound(http.StatusMethodNotAllowed, route+" cannot be "+what+" here")
	}
	routes := &SiteCollectionRoutes{
		Search: cannot("searched"),
		Get:    cannot("read"),
		Create: cannot("created"),
		Update: cannot("updated"),
		Delete: cannot("deleted"),
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

	store.collectionRoutes[section][route] = routes

}

// Collection is the routes of the collection at route in a section, for a route of its own like the summary at /
func (s *SiteStore) Collection(section string, route string) *SiteCollectionRoutes {
	routes, exists := s.collectionRoutes[section][route]
	if !exists {
		panic("no collection at route: " + section + "/" + route)
	}
	return routes
}

// answerNotFound answers as the site does its other errors: a page in a browser, a toast to HTMX and JSON to the API
func answerNotFound(status int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch acceptType := httputils.GetAcceptType(r); acceptType {
		case httputils.AcceptTypeHTML, httputils.AcceptTypeHTMX:
			page := site_templ.SiteError(message)
			if acceptType == httputils.AcceptTypeHTML {
				page = site_templ.FullPage(site_templ.NotFound(message))
			}
			apiutils.WriteStatus(w, acceptType, status)
			page.Render(r.Context(), w)
		default:
			(&apiutils.HTTPError{Status: status, Message: message}).Write(acceptType, w)
		}
	}
}

// routes are those of the collection the request names in a section, or ones answering that there is no such collection
func (s *SiteStore) routes(section string, r *http.Request) *SiteCollectionRoutes {
	name := chi.URLParam(r, "collection")
	if routes, exists := s.collectionRoutes[section][name]; exists {
		return routes
	}
	missing := answerNotFound(http.StatusNotFound, "There is no "+name+" here")
	return &SiteCollectionRoutes{Search: missing, Get: missing, Create: missing, Update: missing, Delete: missing}
}

// Section is the routes of whichever collection in the section a request names, e.g. glucose for /data/glucose
func (s *SiteStore) Section(section string) *SiteCollectionRoutes {
	return &SiteCollectionRoutes{
		Search: func(w http.ResponseWriter, r *http.Request) { s.routes(section, r).Search(w, r) },
		Get:    func(w http.ResponseWriter, r *http.Request) { s.routes(section, r).Get(w, r) },
		Create: func(w http.ResponseWriter, r *http.Request) { s.routes(section, r).Create(w, r) },
		Update: func(w http.ResponseWriter, r *http.Request) { s.routes(section, r).Update(w, r) },
		Delete: func(w http.ResponseWriter, r *http.Request) { s.routes(section, r).Delete(w, r) },
	}
}
