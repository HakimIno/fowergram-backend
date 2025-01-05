package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// GenerateRandomBytes generates a random byte array of specified length
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateRandomCode generates a random numeric code of specified length
func GenerateRandomCode(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)[:length]
}
