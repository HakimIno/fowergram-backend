package security

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// VerifyCost is used for password verification (lower for better performance)
	VerifyCost = 10
	// HashCost is used for password hashing (higher for better security)
	HashCost = bcrypt.DefaultCost
)

// HashPassword creates a bcrypt hash of the password using the higher cost
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), HashCost)
	return string(bytes), err
}

// VerifyPassword checks if the provided password matches the hashed password using the lower cost
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
