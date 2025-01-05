package security

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateRandomCode generates a random numeric code of specified length
func GenerateRandomCode(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
