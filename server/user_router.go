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
	r.Route("/entries", func(r chi.Router) {
		r.Get("/", user.NSGetEntries)
		r.Get(".json", user.NSGetEntries)

		r.Post("/", user.NSCreateEntry)
		r.Post(".json", user.NSCreateEntry)
	})

	// ---- TREATMENTS ----
	r.Route("/treatments", func(r chi.Router) {
		r.Get("/", user.NSGetTreatments)
		r.Get(".json", user.NSGetTreatments)

		r.Post("/", user.NSCreateTreatment)
		r.Post(".json", user.NSCreateTreatment)

	})

	// ---- DEVICE STATUS ----
	r.Route("/devicestatus", func(r chi.Router) {
		r.Get("/", user.NSGetDeviceStatus)
		r.Post("/", user.NSCreateDeviceStatus)
	})

	// ---- PROFILE ----
	r.Route("/profile", func(r chi.Router) {
		r.Get("/", user.NSGetProfiles)
		r.Post("/", user.NSCreateProfile)
	})

	return r
}

func v2Routes() chi.Router {
	r := chi.NewRouter()

	// Entries
	r.Route("/properties", func(r chi.Router) {
		r.Get("/", user.NSGetProperties)

	})

	// TODO token=$token is not a param
	r.Route("/authorization/request", func(r chi.Router) {
		r.Get("/", user.NSGetNightscoutJWTToken)
		r.Post("/", user.NSGetNightscoutJWTToken)
	})

	return r
}

func v3Routes() chi.Router {
	r := chi.NewRouter()

	// Entries (sgv)
	r.Route("/entries", func(r chi.Router) {
		r.Get("/", user.NSGetEntries)
	})

	// Treatments
	r.Route("/treatments", func(r chi.Router) {
		r.Get("/", user.NSGetTreatments)
	})

	return r
}
