// models/user.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleRider  UserRole = "user"
	RoleDriver UserRole = "driver"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone string    `gorm:"uniqueIndex;not null;size:20"`
	Email *string   `gorm:"uniqueIndex;size:255"`
	Name  string    `gorm:"not null;size:100"`

	PasswordHash string `gorm:"not null" json:"-"`
	AvatarURL    *string
	PublicID     *string  `json:"-"`
	Role         UserRole `gorm:"type:user_role;not null" json:"-"`
	IsVerified   bool     `gorm:"not null;default:false"`
	IsActive     bool     `gorm:"not null;default:true"`
	// ReferralCode string     `gorm:"uniqueIndex;not null"`
	// ReferredByID *uuid.UUID `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// ReferredBy *User  `gorm:"foreignKey:ReferredByID"`
	// Referrals  []User `gorm:"foreignKey:ReferredByID"`
}

type PendingUser struct {
	Name         string `json:"name" redis:"name"`
	Email        string `json:"email" redis:"email"`
	HashPassword string `json:"hashPassword" redis:"hashPassword"`
	Phone        string `json:"phone" redis:"phone"`
	OTP          string `json:"otp" redis:"otp"`
	IsVerified   bool   `json:"is_verified" redis:"is_verified"`
	IsActive     bool   `json:"is_active" redis:"is_active"`
}

type UserRegisterInput struct {
	Name     string `json:"name"  validate:"required,min=2,max=100"`
	Password string `json:"password"  validate:"required"`
	Phone    string `json:"phone" validate:"required,e164"`
	Email    string `json:"email" validate:"omitempty,email"`
}

type UserVerifyEmail struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type UserLoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
