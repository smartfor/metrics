package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"net/http"
)

var (
	AuthHeaderName = "HashSHA256"
	CryptoKey      = "AES-Key"
)

func Hash(value []byte) hash.Hash {
	h := sha256.New()
	h.Write(value)
	return h
}

func Sign(value []byte, secret string) hash.Hash {
	return hmac.New(
		func() hash.Hash { return Hash(value) },
		[]byte(secret),
	)
}

func Verify(secret string, hash string, object []byte) bool {
	sign := Sign(object, secret)

	return hmac.Equal(
		sign.Sum(nil),
		[]byte(hash),
	)
}

type HashWriter struct {
	w      http.ResponseWriter
	secret string
}

func (h *HashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *HashWriter) Write(bytes []byte) (int, error) {
	if h.secret != "" {
		sign := Sign(bytes, h.secret)
		h.w.Header().Set(
			AuthHeaderName,
			hex.EncodeToString(sign.Sum(nil)),
		)
	}

	return h.w.Write(bytes)
}

func (h *HashWriter) WriteHeader(statusCode int) {
	h.w.WriteHeader(statusCode)
}

func NewHashWriter(w http.ResponseWriter, secret string) *HashWriter {
	return &HashWriter{
		secret: secret,
		w:      w,
	}
}
