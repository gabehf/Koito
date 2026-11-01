package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID       int32    `json:"id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"` // 'admin' | 'user'
	Password []byte   `json:"-"`
}

type ApiKey struct {
	ID        int32     `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	UserID    int32     `json:"user_id"`
	ProfileID int32     `json:"profile_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID         uuid.UUID
	UserID     int32
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Persistent bool
}
