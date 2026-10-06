package models

import (
	"time"
)

type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

type Permission struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
}