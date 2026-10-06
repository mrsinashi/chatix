package models

import (
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	UserAgent *string   `json:"user_agent,omitempty"`
	IPAddress *string   `json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}