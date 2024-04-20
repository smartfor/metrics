package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/server/utils"
	"net/http"
)

func MakeGetValueHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := core.NewMetricType(chi.URLParam(r, "type"))
		key := chi.URLParam(r, "key")

		v, err := s.Get(metric, key)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(v))
	}
}

func MakeGetValueJSONHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req metrics.Metrics
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		mType := core.NewMetricType(req.MType)

		value, err := s.Get(mType, req.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch mType {
		case core.Counter:
			{
				v, err := utils.CounterFromString(value)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				req.Delta = &v
			}
		case core.Gauge:
			{
				v, err := utils.GaugeFromString(value)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				req.Value = &v
			}
		default:
			http.Error(w, core.ErrUnknownMetricType.Error(), http.StatusInternalServerError)
			return
		}

		if err = json.NewEncoder(w).Encode(req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
