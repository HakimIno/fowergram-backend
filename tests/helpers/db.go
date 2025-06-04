package helpers

import (
	"fowergram/internal/domain"

	"gorm.io/gorm"
)

func CleanupDB(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE users CASCADE")
	db.Exec("TRUNCATE TABLE device_sessions CASCADE")
	db.Exec("TRUNCATE TABLE login_history CASCADE")
	db.Exec("TRUNCATE TABLE auth_codes CASCADE")
	db.Exec("TRUNCATE TABLE account_recovery CASCADE")
}

func SeedTestUser(db *gorm.DB) *domain.User {
	user := &domain.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$...", // pre-hashed password
	}
	db.Create(user)
	return user
}
