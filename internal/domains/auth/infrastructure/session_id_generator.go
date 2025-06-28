package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSessionID generates a cryptographically secure session ID
func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}