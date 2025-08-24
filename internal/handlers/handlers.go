package handlers

import (
	"net/http"

	"nostr-oidc-service/templates"

	"github.com/a-h/templ"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(templates.Login())
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(  templates.Dashboard())
}
