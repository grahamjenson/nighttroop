package v1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v1/handlers"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// NewNightscoutV1Router constructs a router that serves the supported
// Nightscout API v1 endpoints.
func NewNightscoutV1Router(dep nsutils.Dependencies) *chi.Mux {
	r := chi.NewRouter()

	registerStatusRoutes(r, dep)
	registerSearchRoutes(r, dep)
	registerEntryRoutes(r, dep)
	registerTreatmentRoutes(r, dep)
	registerDeviceStatusRoutes(r, dep)
	registerProfileRoutes(r, dep)
	registerSettingRoutes(r, dep)
	registerFoodRoutes(r, dep)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return r
}

func registerStatusRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/status.json", handlers.StatusJSON(dep))
}

func registerSearchRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/verifyauth", handlers.VerifyAuth(dep))
	r.Get("/versions", handlers.Versions())
	r.Get("/echo/{collection}/{spec}.json", dep.Auth.Require("api:entries:read", true, handlers.EchoRoute(dep)))
	r.Get("/slice/{collection}/{field}/{type}/{prefix}.json", dep.Auth.Require("api:entries:read", true, handlers.SliceRoute(dep)))
	r.Get("/times/echo/*", dep.Auth.Require("api:entries:read", true, handlers.TimesEchoRoute(dep)))
	r.Get("/times/*", dep.Auth.Require("api:entries:read", true, handlers.TimesRoute(dep)))
}

func registerEntryRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/entries/current.json", dep.Auth.Require("api:entries:read", true, handlers.EntriesCurrent(dep)))
	r.Get("/entries.json", dep.Auth.Require("api:entries:read", true, handlers.EntriesList(dep)))
	r.Get("/entries/{spec}.json", dep.Auth.Require("api:entries:read", true, handlers.EntriesSpec(dep)))
	r.Post("/entries", dep.Auth.Require("api:entries:create", false, handlers.EntriesCreate(dep, true)))
	r.Post("/entries/", dep.Auth.Require("api:entries:create", false, handlers.EntriesCreate(dep, true)))
	r.Post("/entries.json", dep.Auth.Require("api:entries:create", false, handlers.EntriesCreate(dep, true)))
	r.Post("/entries/preview.json", dep.Auth.Require("api:entries:create", false, handlers.EntriesCreate(dep, false)))
	r.Delete("/entries.json", dep.Auth.Require("api:entries:delete", false, handlers.EntriesDelete(dep)))
	r.Delete("/entries", dep.Auth.Require("api:entries:delete", false, handlers.EntriesDelete(dep)))
	r.Delete("/entries/{spec}", dep.Auth.Require("api:entries:delete", false, handlers.EntriesDelete(dep)))
}

func registerTreatmentRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/treatments", dep.Auth.Require("api:treatments:read", true, handlers.TreatmentsList(dep)))
	r.Get("/treatments.json", dep.Auth.Require("api:treatments:read", true, handlers.TreatmentsList(dep)))
	create := dep.Auth.Require("api:treatments:create", false, handlers.TreatmentsCreate(dep))
	for _, path := range []string{"/treatments", "/treatments/", "/treatments.json"} {
		r.Post(path, create)
		r.Put(path, create)
	}
	r.Delete("/treatments", dep.Auth.Require("api:treatments:delete", false, handlers.TreatmentsDelete(dep)))
	r.Delete("/treatments/", dep.Auth.Require("api:treatments:delete", false, handlers.TreatmentsDelete(dep)))
	r.Delete("/treatments.json", dep.Auth.Require("api:treatments:delete", false, handlers.TreatmentsDelete(dep)))
}

func registerDeviceStatusRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/devicestatus", dep.Auth.Require("api:devicestatus:read", true, handlers.GenericCollectionList(dep, nsutils.Devicestatus)))
	r.Get("/devicestatus.json", dep.Auth.Require("api:devicestatus:read", true, handlers.GenericCollectionList(dep, nsutils.Devicestatus)))
	create := dep.Auth.Require("api:devicestatus:create", false, handlers.GenericCollectionCreate(dep, nsutils.Devicestatus))
	for _, path := range []string{"/devicestatus", "/devicestatus/", "/devicestatus.json"} {
		r.Post(path, create)
		r.Put(path, create)
	}
	r.Delete("/devicestatus", dep.Auth.Require("api:devicestatus:delete", false, handlers.GenericCollectionDelete(dep, nsutils.Devicestatus)))
	r.Delete("/devicestatus/", dep.Auth.Require("api:devicestatus:delete", false, handlers.GenericCollectionDelete(dep, nsutils.Devicestatus)))
	r.Delete("/devicestatus.json", dep.Auth.Require("api:devicestatus:delete", false, handlers.GenericCollectionDelete(dep, nsutils.Devicestatus)))
}

func registerProfileRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/profile", dep.Auth.Require("api:profile:read", true, handlers.ProfileList(dep)))
	r.Get("/profile.json", dep.Auth.Require("api:profile:read", true, handlers.ProfileList(dep)))
	r.Get("/profile/current.json", dep.Auth.Require("api:profile:read", true, handlers.ProfileCurrent(dep)))

	create := dep.Auth.Require("api:profile:create", false, handlers.GenericCollectionCreate(dep, nsutils.Profile))
	for _, path := range []string{"/profile", "/profile/", "/profile.json"} {
		r.Post(path, create)
		r.Put(path, create)
	}

	del := dep.Auth.Require("api:profile:delete", false, handlers.GenericCollectionDelete(dep, nsutils.Profile))
	for _, path := range []string{"/profile", "/profile/", "/profile.json"} {
		r.Delete(path, del)
	}
}

func registerSettingRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	// not supporting any settings changes via API, only logged in Admins can alter settings
}

func registerFoodRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/food", dep.Auth.Require("api:food:read", true, handlers.GenericCollectionList(dep, nsutils.Food)))
	r.Get("/food.json", dep.Auth.Require("api:food:read", true, handlers.GenericCollectionList(dep, nsutils.Food)))

	create := dep.Auth.Require("api:food:create", false, handlers.GenericCollectionCreate(dep, nsutils.Food))
	for _, path := range []string{"/food", "/food/", "/food.json"} {
		r.Post(path, create)
		r.Put(path, create)
	}

	del := dep.Auth.Require("api:food:delete", false, handlers.GenericCollectionDelete(dep, nsutils.Food))
	for _, path := range []string{"/food", "/food/", "/food.json"} {
		r.Delete(path, del)
	}
}
