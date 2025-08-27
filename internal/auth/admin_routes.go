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
	"github.com/lescuer97/nostr-oicd/internal/ui"
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
			// render OOB error fragment into the stable snackbar
			_ = ui.RenderSnackbar(r.Context(), w, "invalid form", "error", "5s")
			w.WriteHeader(http.StatusOK)
			return
		}
		npub := r.FormValue("npub")
		if npub == "" {
			_ = ui.RenderSnackbar(r.Context(), w, "npub is required", "error", "5s")
			w.WriteHeader(http.StatusOK)
			slog.Warn("admin_add_user_missing_npub", "admin", adminPub, "remote", r.RemoteAddr)
			return
		}

		// convert npub to hex public key — npub is bech32 (nip19) encoded pubkey
		pubHex, err := decodeNpubToHex(npub)
		if err != nil {
			_ = ui.RenderSnackbar(r.Context(), w, fmt.Sprintf("invalid npub: %v", err), "error", "5s")
			w.WriteHeader(http.StatusOK)
			slog.Error("admin_add_user_invalid_npub", "admin", adminPub, "remote", r.RemoteAddr, "npub", npub, "error", err.Error())
			return
		}

		ctx := r.Context()
		// try to ensure user (EnsureUser will create if missing)
		id, err := models.EnsureUser(ctx, db, pubHex)
		if err != nil {
			_ = ui.RenderSnackbar(r.Context(), w, fmt.Sprintf("failed to ensure user: %v", err), "error", "5s")
			w.WriteHeader(http.StatusOK)
			slog.Error("admin_add_user_db_error", "admin", adminPub, "remote", r.RemoteAddr, "pubHex", pubHex, "error", err.Error())
			return
		}

		// success: show a success snackbar
		_ = ui.RenderSnackbar(r.Context(), w, fmt.Sprintf("user added (id=%d)", id), "success", "5s")
		w.WriteHeader(http.StatusOK)
		slog.Info("admin_add_user_success", "admin", adminPub, "remote", r.RemoteAddr, "pubHex", pubHex, "user_id", id)
		return
	})).ServeHTTP)
}
