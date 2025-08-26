package middleware

import (
	"net/http"

	"github.com/lescuer97/nostr-oicd/internal/models"
)

// AdminOnly ensures the current user in context is an admin. Requires AuthMiddleware ran before.
func AdminOnly() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := r.Context().Value(ContextUserKey)
			if u == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			user, ok := u.(*models.User)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !user.IsAdmin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
