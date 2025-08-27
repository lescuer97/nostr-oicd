package ui

import (
	"context"
	"net/http"

	"github.com/lescuer97/nostr-oicd/templates/fragments"
)

// RenderSnackbar renders the Snackbar fragment and sets HX-Retarget so HTMX replaces
// the inner HTML of the stable snackbar element (#htmx-snackbar).
func RenderSnackbar(ctx context.Context, w http.ResponseWriter, msg, severity, remove string) error {
	w.Header().Set("HX-Retarget", "#htmx-snackbar")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return fragments.Snackbar(msg, severity, remove).Render(ctx, w)
}
