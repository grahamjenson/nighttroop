package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
)

func NewUserRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	AddStandardMiddleware(r, config)

	r.Use(user.UserRateLimitMiddleware(config))

	static.AddSatticRoutes(r)
	// Rate Limiter is before auth because auth can be expencive

	// Publically availabl (if the person knows the subdomain) endpoints
	r.Get("/login", user.GETLoginPageHandler())
	r.Post("/login", user.POSTLogin(controlDB, config))
	r.Post("/login_totp", user.POSTLoginTOTP(controlDB, config))

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(user.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(user.UserDBMiddleware(controlDB))          // Adds User DB and model to the context

		r.Get("/", user.GetDashboardHandler(controlDB, config))
		r.Post("/logout", user.POSTLogoutHandler(config))

		//// TODO
		// Get Devices probably with pagination returns htmx, include generated tokens and last seen
		// Delete Device
		// Change Device Passowrd, role TOTP
		// Get Device Token

	})

	return r
}

func NewNightscoutRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	AddStandardMiddleware(r, config)
	r.Use(user.UserRateLimitMiddleware(config))
	// No unauthed endpoints in Nightscout
	// It is encouraged that users update their auth to JWT tokens for speed and security though
	r.Use(user.NightscoutAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth

	// Versioned APIs
	r.Mount("/api/v1", v1Routes())
	r.Mount("/api/v2", v2Routes())
	r.Mount("/api/v3", v3Routes())

	return r
}

func v1Routes() chi.Router {
	r := chi.NewRouter()

	// ---- ENTRIES (sgv) ----
	r.Get("/entries", user.NSGetEntries)
	r.Get("/entries.json", user.NSGetEntries)
	r.Post("/entries", user.NSCreateEntry)
	r.Post("/entries.json", user.NSCreateEntry)

	// ---- TREATMENTS ----
	r.Get("/treatments", user.NSGetTreatments)
	r.Get("/treatments.json", user.NSGetTreatments)
	r.Post("/treatments", user.NSCreateTreatment)
	r.Post("/treatments.json", user.NSCreateTreatment)

	// ---- DEVICE STATUS ----
	r.Get("/devicestatus", user.NSGetDeviceStatus)
	r.Get("/devicestatus.json", user.NSGetDeviceStatus)
	r.Post("/devicestatus", user.NSCreateDeviceStatus)
	r.Post("/devicestatus.json", user.NSCreateDeviceStatus)

	// ---- PROFILE ----
	r.Get("/profile", user.NSGetProfiles)
	r.Get("/profile.json", user.NSGetProfiles)
	r.Post("/profile", user.NSCreateProfile)
	r.Post("/profile.json", user.NSCreateProfile)

	return r
}

func v2Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/properties", user.NSGetProperties)
	r.Get("/properties.json", user.NSGetProperties)

	// TODO token=$token is not a param
	r.Get("/authorization/request", user.NSGetNightscoutJWTToken)
	r.Post("/authorization/request", user.NSGetNightscoutJWTToken)

	return r
}

func v3Routes() chi.Router {
	r := chi.NewRouter()

	// Entries (sgv)
	r.Get("/entries", user.NSGetEntries)

	// Treatments
	r.Get("/treatments", user.NSGetTreatments)

	return r
}
