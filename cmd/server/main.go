package main

import (
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/storage"
	"net/http"
	"strings"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metricStorage := storage.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", createUpdateHandler(metricStorage))

	return http.ListenAndServe(`:8080`, mux)
}

func createUpdateHandler(s core.Storage) func(http.ResponseWriter, *http.Request) {
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
			if err.Type == core.UnknownMetricType || err.Type == core.BadMetricValue {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
