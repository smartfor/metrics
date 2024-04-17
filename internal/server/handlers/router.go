package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
)

func Router(s core.Storage) chi.Router {
	r := chi.NewRouter()

	r.Get("/", MakeGetMetricsPageHandler(s))
	r.Post("/update/{type}/{key}/{value}", MakeUpdateHandler(s))
	r.Get("/value/{type}/{key}", MakeGetValueHandler(s))

	return r
}
