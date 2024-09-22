package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/server/utils"
)

// MakeGetValueHandler создает хендлер для получения значения метрики в формате строки
func MakeGetValueHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := core.NewMetricType(chi.URLParam(r, "type"))
		key := chi.URLParam(r, "key")

		v, err := s.Get(r.Context(), key, metric)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(v))
	}
}

// MakeGetValueHandler создает хендлер для получения значения метрики в формате JSON
func MakeGetValueJSONHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req metrics.Metrics
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, err, http.StatusBadRequest)
			return
		}

		mType := core.NewMetricType(req.MType)

		value, err := s.Get(r.Context(), req.ID, mType)
		if err != nil {
			utils.WriteError(w, err, http.StatusNotFound)
			return
		}

		switch mType {
		case core.Counter:
			{
				v, err := utils.CounterFromString(value)
				if err != nil {
					utils.WriteError(w, err, http.StatusInternalServerError)
					return
				}
				req.Delta = &v
			}
		case core.Gauge:
			{
				v, err := utils.GaugeFromString(value)
				if err != nil {
					utils.WriteError(w, err, http.StatusInternalServerError)
					return
				}
				req.Value = &v
			}
		default:
			utils.WriteError(w, core.ErrUnknownMetricType, http.StatusBadRequest)
			return
		}

		if err = json.NewEncoder(w).Encode(req); err != nil {
			utils.WriteError(w, err, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
