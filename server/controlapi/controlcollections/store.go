package controlcollections

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ControlCollectionRoutes struct {
	Search http.HandlerFunc
	Get    http.HandlerFunc
	Update http.HandlerFunc
}

type ControlStore struct {
	// collectionRoutes are the collections by route, e.g. sites for /sites
	collectionRoutes map[string]*ControlCollectionRoutes
}

func NewControlStore() *ControlStore {
	store := &ControlStore{collectionRoutes: make(map[string]*ControlCollectionRoutes)}

	store.registerCollection(&SitesCollection{})
	store.registerCollection(&JobStatsCollection{})
	store.registerCollection(&JobsCollection{})
	store.registerCollection(&SyncSourcesCollection{})

	return store
}

func (store *ControlStore) registerCollection[R, M any](collection CollectionBase[R, M]) {
	route := collection.GetRoute()
	if _, exists := store.collectionRoutes[route]; exists {
		panic("collection route already registered: " + route)
	}

	cannot := func(what string) http.HandlerFunc {
		message := route + " cannot be " + what + " here"
		return func(w http.ResponseWriter, r *http.Request) {
			(&apiutils.HTTPError{Status: http.StatusMethodNotAllowed, Message: message}).Write(httputils.GetAcceptType(r), w)
		}
	}
	routes := &ControlCollectionRoutes{Search: cannot("searched"), Get: cannot("read"), Update: cannot("updated")}

	added := false
	if val, ok := collection.(CollectionSearch[R, M]); ok {
		routes.Search = SearchHandler(val)
		added = true
	}
	if val, ok := collection.(CollectionGet[R, M]); ok {
		routes.Get = GetHandler(val)
		added = true
	}
	if val, ok := collection.(CollectionUpdate[R, M]); ok {
		routes.Update = UpdateHandler(val)
		added = true
	}
	if !added {
		panic("collection does not implement any known interfaces: " + collection.GetName())
	}

	store.collectionRoutes[route] = routes
}

// Collection is the routes of the collection at route, for a route of its own like the sites at /
func (store *ControlStore) Collection(route string) *ControlCollectionRoutes {
	routes, exists := store.collectionRoutes[route]
	if !exists {
		panic("no collection at route: " + route)
	}
	return routes
}

// Routes adds each collection's routes at its own path, e.g. /sites and /sites/{id}
func (store *ControlStore) Routes(r chi.Router) {
	for route, routes := range store.collectionRoutes {
		r.Get("/"+route, routes.Search)
		r.Get("/"+route+"/{id}", routes.Get)
		r.Post("/"+route+"/{id}", routes.Update)
	}
}
