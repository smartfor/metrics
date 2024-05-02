package utils

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, error error, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": error.Error(),
	})
}
