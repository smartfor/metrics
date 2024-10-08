package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/server/utils"
)

// MakeUpdateHandler создает хендлер для обновления метрики в строковом формате
func MakeUpdateHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := core.NewMetricType(chi.URLParam(r, "type"))
		key := chi.URLParam(r, "key")
		value := chi.URLParam(r, "value")

		err := s.Set(r.Context(), key, value, metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// MakeUpdateJSONHandler создает хендлер для обновления метрики в формате JSON
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
				err := s.Set(r.Context(), req.ID, utils.CounterAsString(*req.Delta), mType)
				if err != nil {
					utils.WriteError(w, err, http.StatusBadRequest)
					return
				}

				newValue, err := s.Get(r.Context(), req.ID, mType)
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
				err := s.Set(r.Context(), req.ID, utils.GaugeAsString(*req.Value), mType)
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

func MakeBatchUpdateJSONHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		defer r.Body.Close()

		var req []metrics.Metrics
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, err, http.StatusBadRequest)
			return
		}

		gauges := make(map[string]float64)
		counters := make(map[string]int64)

		for _, m := range req {
			switch core.NewMetricType(m.MType) {
			case core.Gauge:
				gauges[m.ID] = *m.Value
			case core.Counter:
				v, ok := counters[m.ID]
				if !ok {
					v = 0
				}
				counters[m.ID] = v + *m.Delta
			default:
				utils.WriteError(w, core.ErrUnknownMetricType, http.StatusBadRequest)
				return
			}
		}

		batch := core.NewBaseMetricStorageWithValues(gauges, counters)
		if err := s.SetBatch(r.Context(), batch); err != nil {
			utils.WriteError(w, err, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
