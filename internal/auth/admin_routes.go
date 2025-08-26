package auth

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lescuer97/nostr-oicd/internal/config"
	"github.com/lescuer97/nostr-oicd/internal/middleware"
	"github.com/lescuer97/nostr-oicd/internal/models"
	"github.com/lescuer97/nostr-oicd/templates/fragments"
)

// Register additional admin routes onto router r. Requires middleware.AuthMiddleware used earlier.
func RegisterAdminRoutes(r chi.Router, cfg *config.Config, db *sql.DB) {
	// show add user form (HTMX fragment)
	r.HandleFunc("/admin/users/new", middleware.AdminOnly()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// audit log: admin viewed add-user form
		if u := r.Context().Value(middleware.ContextUserKey); u != nil {
			if user, ok := u.(*models.User); ok {
				slog.Info("admin_view_add_user_form", "admin", user.PublicKey, "remote", r.RemoteAddr)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = fragments.AdminAddUserForm().Render(r.Context(), w)
	})).ServeHTTP)

	// POST handler to create user by npub
	r.HandleFunc("/admin/users/add", middleware.AdminOnly()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// audit context: admin user
		var adminPub string
		if u := r.Context().Value(middleware.ContextUserKey); u != nil {
			if user, ok := u.(*models.User); ok {
				adminPub = user.PublicKey
			}
		}

		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = fragments.AdminMessageError("invalid form").Render(r.Context(), w)
			slog.Error("admin_add_user_failed", "admin", adminPub, "remote", r.RemoteAddr, "error", err.Error())
			return
		}
		npub := r.FormValue("npub")
		if npub == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = fragments.AdminMessageError("npub is required").Render(r.Context(), w)
			slog.Warn("admin_add_user_missing_npub", "admin", adminPub, "remote", r.RemoteAddr)
			return
		}
		// convert npub to hex public key — npub is bech32 (nip19) encoded pubkey
		pubHex, err := decodeNpubToHex(npub)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = fragments.AdminMessageError(fmt.Sprintf("invalid npub: %v", err)).Render(r.Context(), w)
			slog.Error("admin_add_user_invalid_npub", "admin", adminPub, "remote", r.RemoteAddr, "npub", npub, "error", err.Error())
			return
		}
		ctx := r.Context()
		// try to ensure user (EnsureUser will create if missing)
		id, err := models.EnsureUser(ctx, db, pubHex)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = fragments.AdminMessageError(fmt.Sprintf("failed to ensure user: %v", err)).Render(r.Context(), w)
			slog.Error("admin_add_user_db_error", "admin", adminPub, "remote", r.RemoteAddr, "pubHex", pubHex, "error", err.Error())
			return
		}
		_ = fragments.AdminMessageSuccess(fmt.Sprintf("user added (id=%d)", id)).Render(r.Context(), w)
		slog.Info("admin_add_user_success", "admin", adminPub, "remote", r.RemoteAddr, "pubHex", pubHex, "user_id", id)
	})).ServeHTTP)
}
