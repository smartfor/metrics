package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/middlewares"
	"go.uber.org/zap"
)

func Router(s core.Storage, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.MakeLogger(logger))

	//logMiddleware := middlewares.MakeLogger(logger)

	r.Get("/", MakeGetMetricsPageHandler(s))
	r.Post("/update/{type}/{key}/{value}", MakeUpdateHandler(s))
	r.Post("/update/", MakeUpdateJsonHandler(s))
	r.Get("/value/{type}/{key}", MakeGetValueHandler(s))
	r.Post("/value/", MakeGetValueJsonHandler(s))

	return r
}
