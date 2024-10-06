package utils

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, error error, statusCode int) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(map[string]string{
		"error": error.Error(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
