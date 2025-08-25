package auth

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lescuer97/nostr-oicd/internal/config"
	"github.com/lescuer97/nostr-oicd/internal/middleware"
)

// RegisterRoutes registers auth routes on the provided router and passes the DB to handlers.
func RegisterRoutes(r chi.Router, cfg *config.Config, db *sql.DB) {
	r.Post("/api/auth/challenge", ChallengeHandler)
	// login expects cfg+DB for session creation
	r.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) { LoginHandler(cfg, db, w, r) })
	// TODO: add /signup, /logout, /status

	// Example protected route that requires authentication (for testing)
	r.With(middleware.AuthMiddleware(cfg, db)).Get("/api/auth/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\":\"ok\"}"))
	})
}
