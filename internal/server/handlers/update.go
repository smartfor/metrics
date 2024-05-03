package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/server/utils"
	"net/http"
)

func MakeUpdateHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := core.NewMetricType(chi.URLParam(r, "type"))
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

func MakeUpdateJSONHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		defer r.Body.Close()

		var req metrics.Metrics
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, err, http.StatusBadRequest)
			return
		}

		mType := core.NewMetricType(req.MType)

		switch mType {
		case core.Counter:
			{
				err := s.Set(mType, req.ID, utils.CounterAsString(*req.Delta))
				if err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}

				newValue, err := s.Get(mType, req.ID)
				if err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}

				counter, err := utils.CounterFromString(newValue)
				if err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}

				*req.Delta = counter
				if err = json.NewEncoder(w).Encode(req); err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}
			}
		case core.Gauge:
			{
				err := s.Set(mType, req.ID, utils.GaugeAsString(*req.Value))
				if err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}

				if err = json.NewEncoder(w).Encode(req); err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}
			}
		default:
			utils.WriteError(w, core.ErrUnknownMetricType, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
