package auth

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lescuer97/nostr-oicd/internal/config"
	"github.com/lescuer97/nostr-oicd/internal/middleware"
	"github.com/lescuer97/nostr-oicd/internal/models"
	pages "github.com/lescuer97/nostr-oicd/templates/pages"
)

// RegisterRoutes registers auth routes on the provided router and passes the DB to handlers.
func RegisterRoutes(r chi.Router, cfg *config.Config, db *sql.DB) {
	// Configure rate limiters for auth endpoints
	// login: 5 requests per minute with burst 10
	loginLimiter := middleware.RateLimitMiddleware(middleware.PerMinute(5), 10)
	// challenge endpoints: 20 requests per minute with burst 40
	challengeLimiter := middleware.RateLimitMiddleware(middleware.PerMinute(20), 40)

	// Allow GET for HTMX fragment load and POST for programmatic flows
	r.With(challengeLimiter).Get("/api/auth/challenge", ChallengeHandler)
	r.With(challengeLimiter).Post("/api/auth/challenge", ChallengeHandler)

	// login expects cfg+DB for session creation
	r.With(loginLimiter).Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) { LoginHandler(cfg, db, w, r) })
	// TODO: add /signup, /status

	// Logout route (protected) — POST
	r.With(middleware.AuthMiddleware(cfg, db)).Post("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(cfg, db, w, r)
	})

	// Dashboard route (requires authentication)
	r.With(middleware.AuthMiddleware(cfg, db)).Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		// get user from context
		u := r.Context().Value(middleware.ContextUserKey)
		if u == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		user := u.(*models.User)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// render dashboard with admin flag
		if err := pages.DashboardPage(user.PublicKey, user.IsAdmin).Render(r.Context(), w); err != nil {
			http.Error(w, "failed to render", http.StatusInternalServerError)
		}
	})

	// admin routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg, db))
		r.Use(middleware.AdminOnly())
		RegisterAdminRoutes(r, cfg, db)
	})
}
