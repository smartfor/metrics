package handlers

import (
	"fmt"
	"github.com/smartfor/metrics/internal/core"
	"log"
	"net/http"
	"sort"
)

func MakeGetMetricsPageHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := s.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var out = "<ul>"
		for _, k := range keys {
			out += fmt.Sprintf("<li>%s : %s</li>", k, m[k])
		}
		out += "</ul>"

		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(out)); err != nil {
			log.Printf("Error writing response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
