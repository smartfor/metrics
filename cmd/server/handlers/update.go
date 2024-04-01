package handlers

import (
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"net/http"
	"strings"
)

func MakeUpdateHandler(s core.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		split := strings.Split(r.URL.Path, "/")[2:]
		if len(split) < 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metricType := split[0]
		metric := metrics.NewMetricType(metricType)

		key := split[1]
		value := split[2]

		err := s.Set(metric, key, value)
		if err != nil {
			//if err.Type == core.UnknownMetricType || err.Type == core.BadMetricValue {
			//	w.WriteHeader(http.StatusBadRequest)
			//	return
			//}

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
