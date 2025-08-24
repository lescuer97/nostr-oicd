package db

import "time"

type User struct {
	ID           int        `json:"id"`
	PublicKeyHex string     `json:"public_key_hex"`
	DisplayName  string     `json:"display_name,omitempty"`
	Active       bool       `json:"active"`
	IsAdmin      bool       `json:"is_admin"`
	CreatedAt    time.Time  `json:"creation_timestamp"`
	UpdatedAt    time.Time  `json:"updated_timestamp"`
	DeletedAt    *time.Time `json:"deletion_timestamp,omitempty"`
}

type Session struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"`
	UserID    int       `json:"user_id"`
	Client    string    `json:"client"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"creation_timestamp"`
	ExpiresAt time.Time `json:"expiry_timestamp"`
}
