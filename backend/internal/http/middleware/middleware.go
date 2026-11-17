package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

func Recover(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func Logger(next http.Handler) http.Handler {
	return middleware.Logger(next)
}

func CORS(origins string) func(http.Handler) http.Handler {
	allowed := strings.Split(origins, ",")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			ok := false
			for _, o := range allowed {
				o = strings.TrimSpace(o)
				if o == "*" || o == origin {
					ok = true
					break
				}
			}
			if ok {
				if origins == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

var _ = slog.Info
