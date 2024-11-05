package middlewares

import (
	"log"
	"net/http"

	"github.com/smartfor/metrics/internal/ip"
)

func MakeInSubnetMiddleware(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := ip.ExtractXRealIP(r.Header)
			log.Println("RealIP: ", realIP)

			if !ip.InSubnet(realIP, trustedSubnet) {
				log.Println("Unauthorized request from", realIP)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

}
