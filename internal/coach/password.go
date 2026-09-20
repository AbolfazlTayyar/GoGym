package coach

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword bcrypt-hashes a plaintext password for storage.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("coach: failed to hash password: %w", err)
	}
	return string(hash), nil
}

// verifyPassword reports whether password matches hash.
func verifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
