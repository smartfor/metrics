package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"net/http"
)

func MakeGetValueHandler(s core.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := metrics.NewMetricType(chi.URLParam(r, "type"))
		key := chi.URLParam(r, "key")

		v, err := s.Get(metric, key)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(v))
	}
}
