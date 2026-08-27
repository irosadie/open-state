package entities

import "time"

type UserRole string
type UserStatus string

const (
	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"

	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         UserRole
	Status       UserStatus
	Photo        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
