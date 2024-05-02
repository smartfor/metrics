package handlers

import (
	"context"
	"github.com/smartfor/metrics/internal/server/storage"
	"net/http"
)

func MakePingHandler(dbStorage *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if dbStorage == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := dbStorage.Pool.Ping(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
