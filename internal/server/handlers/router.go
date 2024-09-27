// Пакет с хендлерами сервера
package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/middlewares"
	"go.uber.org/zap"
)

// Router создает роутер сервера со всем обработчиками ендпоинтов включая ендпоинты профилирования
func Router(s core.Storage, logger *zap.Logger, secret string) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.MakeLoggerMiddleware(logger))

	r.Get("/ping", MakePingHandler(s))

	r.Get("/", MakeGetMetricsPageHandler(s))

	r.Post("/value/", MakeGetValueJSONHandler(s))
	r.Get("/value/{type}/{key}", MakeGetValueHandler(s))

	r.Mount("/debug", middleware.Profiler())

	r.Group(func(r chi.Router) {
		if secret != "" {
			r.Use(middlewares.MakeAuthMiddleware(secret))
		}
		r.Post("/updates/", MakeBatchUpdateJSONHandler(s))
		r.Post("/update/", MakeUpdateJSONHandler(s))
		r.Post("/update/{type}/{key}/{value}", MakeUpdateHandler(s))
	})

	return r
}
