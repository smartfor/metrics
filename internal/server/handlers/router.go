package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/middlewares"
	"go.uber.org/zap"
)

func Router(s core.Storage, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.MakeLoggerMiddleware(logger))
	r.Use(middlewares.GzipMiddleware)

	r.Get("/", MakeGetMetricsPageHandler(s))
	r.Post("/update/{type}/{key}/{value}", MakeUpdateHandler(s))
	r.Get("/value/{type}/{key}", MakeGetValueHandler(s))

	r.Group(func(r chi.Router) {
		r.Post("/update/", MakeUpdateJSONHandler(s))
		r.Post("/value/", MakeGetValueJSONHandler(s))
	})

	return r
}
