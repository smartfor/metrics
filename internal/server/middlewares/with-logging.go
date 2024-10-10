package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		body   interface{}
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)

	r.responseData.size += size
	r.responseData.body = string(b)

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// MakeLoggerMiddleware - middleware для логирования основной информации запросов и ответов
func MakeLoggerMiddleware(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			sugar := logger.Sugar()

			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)
			sugar.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"request", string(bodyBytes),
				"duration", duration,
				"status", responseData.status,
				"size", responseData.size,
				"response", responseData.body,
			)
		}

		return http.HandlerFunc(fn)
	}

}
