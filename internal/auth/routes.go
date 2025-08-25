package auth

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers auth routes on the provided router and passes the DB to handlers.
func RegisterRoutes(r chi.Router, db *sql.DB) {
	r.Post("/api/auth/challenge", ChallengeHandler)
	// login expects DB for session creation
	r.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) { LoginHandler(db, w, r) })
	// TODO: add /signup, /logout, /status
}
