// Package middlewares содержит функции для обработки запросов.
package middlewares

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/smartfor/metrics/internal/utils"
)

// MakeAuthMiddleware - middleware для проверки ключа аутентификации
func MakeAuthMiddleware(secret string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			hexHash := r.Header.Get(utils.AuthHeaderName)
			if hexHash == "" {
				// http.Error(w, "Empty Hash", http.StatusBadRequest)
				h.ServeHTTP(w, r)
				return
			}

			hashBytes, err := hex.DecodeString(hexHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !utils.Verify(secret, string(hashBytes), bodyBytes) {
				http.Error(w, "Invalid Hash", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			hw := utils.NewHashWriter(w, secret)
			h.ServeHTTP(hw, r)
		}

		return http.HandlerFunc(fn)
	}
}
