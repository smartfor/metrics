package handlers

import (
	"fmt"
	"github.com/smartfor/metrics/internal/core"
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

		w.Write([]byte(out))
	}
}
