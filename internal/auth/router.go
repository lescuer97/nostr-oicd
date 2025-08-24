package auth

import (
	"database/sql"
	"github.com/gorilla/mux"
)

func InitAuthRoutes(db *sql.DB, r *mux.Router) {
	r.HandleFunc("/auth/challenge", ChallengeHandler(db)).Methods("GET")
	r.HandleFunc("/auth/verify", VerifyHandler(db)).Methods("POST")
}
