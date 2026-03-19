package router

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/db"
)

func NewControlRouter(controlDB *db.ControlDB) chi.Router {
	r := chi.NewRouter()

	return r
}
