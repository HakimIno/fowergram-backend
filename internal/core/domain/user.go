package domain

import "time"

type User struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	Username            string     `json:"username" gorm:"unique"`
	Email               string     `json:"email" gorm:"unique"`
	PasswordHash        string     `json:"-"`
	RecoveryEmail       string     `json:"recovery_email,omitempty"`
	IsEmailVerified     bool       `json:"is_email_verified"`
	FailedLoginAttempts int        `json:"-"`
	LastFailedLogin     *time.Time `json:"-"`
	AccountLockedUntil  *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
