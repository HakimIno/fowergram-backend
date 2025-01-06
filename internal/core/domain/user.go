package domain

import "time"

type User struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	Username            string     `json:"username" gorm:"unique;not null"`
	Email               string     `json:"email" gorm:"unique;not null"`
	PasswordHash        string     `json:"-" gorm:"not null"`
	IsEmailVerified     bool       `json:"is_email_verified" gorm:"default:false"`
	RecoveryEmail       string     `json:"recovery_email,omitempty"`
	FailedLoginAttempts int        `json:"-" gorm:"default:0"`
	LastFailedLogin     *time.Time `json:"-"`
	AccountLockedUntil  *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
