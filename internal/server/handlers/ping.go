package handlers

import (
	"github.com/smartfor/metrics/internal/core"
	"net/http"
)

func MakePingHandler(s core.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
