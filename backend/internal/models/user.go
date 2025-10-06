// path: backend/internal/models/user.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

type User struct {
	ID                         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email                      string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash               string     `gorm:"not null" json:"-"`
	FirstName                  string     `json:"first_name"`
	LastName                   string     `json:"last_name"`
	Role                       UserRole   `gorm:"default:user" json:"role"`
	TeamID                     *uuid.UUID `gorm:"type:uuid" json:"team_id"`
	EmailVerified              bool       `gorm:"default:false" json:"email_verified"`
	VerificationToken          *string    `json:"-"`
	VerificationTokenExpiresAt *time.Time `json:"-"`
	ResetToken                 *string    `json:"-"`
	ResetTokenExpiresAt        *time.Time `json:"-"`
	LastLoginAt                *time.Time `json:"last_login_at"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
	DeletedAt                  *time.Time `gorm:"index" json:"-"`
}

type Team struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string     `gorm:"not null" json:"name"`
	Plan      string     `gorm:"default:free" json:"plan"`
	OwnerID   *uuid.UUID `gorm:"type:uuid" json:"owner_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TokenHash string    `gorm:"not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

func (Team) TableName() string {
	return "teams"
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
