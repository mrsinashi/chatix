package models

import (
	"time"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusBlocked  UserStatus = "blocked"
	UserStatusArchived UserStatus = "archived"
)

type UserKind string

const (
	UserKindPerson UserKind = "person"
	UserKindRoom   UserKind = "room"
	UserKindRole   UserKind = "role"
)

type User struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	UsernameNormalized string     `json:"-"`
	Kind               UserKind   `json:"kind"`
	PasswordHash       string     `json:"-"`
	DisplayName        string     `json:"display_name"`
	RoomText           *string    `json:"room_text,omitempty"`
	RoleTitle          *string    `json:"role_title,omitempty"`
	Email              *string    `json:"email,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	DepartmentID       *string    `json:"department_id,omitempty"`
	AvatarFileID       *string    `json:"avatar_file_id,omitempty"`
	MustChangePassword bool       `json:"must_change_password"`
	Status             UserStatus `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastSeenAt         *time.Time `json:"last_seen_at,omitempty"`
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