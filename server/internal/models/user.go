package models

import (
	"time"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

type User struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	UsernameNormalized string     `json:"-"`
	PasswordHash       string     `json:"-"`
	DisplayName        string     `json:"display_name"`
	Email              *string    `json:"email,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	DepartmentID       *string    `json:"department_id,omitempty"`
	Status             UserStatus `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
}

type UserCreate struct {
	Username    string     `json:"username"`
	Password    string     `json:"password"`
	DisplayName string     `json:"display_name"`
	Email       *string    `json:"email,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	DepartmentID *string   `json:"department_id,omitempty"`
	Status      UserStatus `json:"status"`
}

type UserUpdate struct {
	DisplayName *string     `json:"display_name,omitempty"`
	Email       *string     `json:"email,omitempty"`
	Phone       *string     `json:"phone,omitempty"`
	DepartmentID *string    `json:"department_id,omitempty"`
	Status      *UserStatus `json:"status,omitempty"`
}