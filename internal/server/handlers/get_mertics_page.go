package handlers

import (
	"fmt"
	"github.com/smartfor/metrics/internal/core"
	"log"
	"net/http"
)

func MakeGetMetricsPageHandler(s core.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := s.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		var out = ""
		for k, v := range m {
			out += fmt.Sprintf("%s : %s\n", k, v)
		}

		if _, err := w.Write([]byte(out)); err != nil {
			log.Printf("Error writing response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
