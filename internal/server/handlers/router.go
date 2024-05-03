package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/middlewares"
	"go.uber.org/zap"
)

func Router(s core.Storage, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()

	//r.Use(middleware.Logger)
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.MakeLoggerMiddleware(logger))

	r.Get("/ping", MakePingHandler(s))

	r.Get("/", MakeGetMetricsPageHandler(s))
	r.Post("/update/", MakeUpdateJSONHandler(s))
	r.Post("/updates/", MakeBatchUpdateJSONHandler(s))
	r.Post("/value/", MakeGetValueJSONHandler(s))
	r.Post("/update/{type}/{key}/{value}", MakeUpdateHandler(s))
	r.Get("/value/{type}/{key}", MakeGetValueHandler(s))

	return r
}
