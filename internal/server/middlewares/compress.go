package middlewares

import (
	"net/http"
	"strings"

	"github.com/smartfor/metrics/internal/utils"
)

func GzipMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip && isCompressType(&r.Header) {
			ow.Header().Set("Content-Encoding", "gzip")

			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := utils.NewCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := utils.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		//передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(fn)
}

func isCompressType(header *http.Header) bool {
	if header.Get("Content-Type") == "application/json" {
		return true
	}

	acceptHeader := header.Get("Accept")
	if strings.Contains(acceptHeader, "text/html") ||
		strings.Contains(acceptHeader, "html/text") {
		return true
	}

	return false
}
