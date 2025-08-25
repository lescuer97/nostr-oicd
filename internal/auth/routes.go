package auth

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers auth routes on the provided router.
func RegisterRoutes(r chi.Router) {
	r.Post("/api/auth/challenge", ChallengeHandler)
	// TODO: add /api/auth/login, /signup, /logout, /status
}
