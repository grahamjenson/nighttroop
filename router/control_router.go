package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/pquerna/otp/totp"
	"nighttroop.t1d.nz/data/db"
)

func NewControlRouter(controlDB *db.ControlDB) chi.Router {
	r := chi.NewRouter()

	totp.Validate("", "") // TODO
	return r
}
