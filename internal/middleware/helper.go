package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ApplyRateLimit sets a limiter middleware on the router for a given path and method
func ApplyRateLimit(r chi.Router, method, path string, middlewareFunc func(next http.Handler) http.Handler) {
	switch method {
	case "GET":
		r.With(middlewareFunc).Get(path, chi.Route{}.(func(w http.ResponseWriter, r *http.Request)))
	default:
		// No-op: this helper is minimal and not used broadly
	}
}
