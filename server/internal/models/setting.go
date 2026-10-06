package models

import (
	"encoding/json"
	"time"
)

type Setting struct {
	ID        string          `json:"id"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UserID    *string         `json:"user_id,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}