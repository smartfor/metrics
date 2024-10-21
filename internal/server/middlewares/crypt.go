package middlewares

import (
	"bytes"
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/smartfor/metrics/internal/utils"
)

func MakeCryptoMiddleware(cryptoKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println("Error reading request body:", err)
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			encryptedKey := r.Header.Get(utils.CryptoKey)
			if encryptedKey == "" {
				log.Println("No key found in request header")
				http.Error(w, "No key found", http.StatusBadRequest)
				return
			}
			key, err := hex.DecodeString(encryptedKey)
			if err != nil {
				log.Println("Error decoding key:", err)
				http.Error(w, "Failed to decode key", http.StatusBadRequest)
				return
			}

			// Decrypt the message using the private key
			decodedMessage, err := utils.DecryptWithPrivateKey(data, key, cryptoKey)
			if err != nil {
				log.Println("Error decrypting message:", err)
				http.Error(w, "Failed to decrypt message", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decodedMessage))
			next.ServeHTTP(w, r)
		})
	}
}
