package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         string          `json:"id"`
	UserID     *string         `json:"user_id,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AuditLogCreate struct {
	UserID     *string         `json:"user_id,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details"`
	IPAddress  *string         `json:"ip_address,omitempty"`
}