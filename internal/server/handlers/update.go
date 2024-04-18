package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"net/http"
)

func MakeUpdateHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := metrics.NewMetricType(chi.URLParam(r, "type"))
		key := chi.URLParam(r, "key")
		value := chi.URLParam(r, "value")

		err := s.Set(metric, key, value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
