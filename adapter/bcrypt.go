// Package adapter provides infrastructure implementations shared across
// multiple driven adapters (e.g. password hashing).
package adapter

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// BcryptHasher implements domain.PasswordHasher using bcrypt.
type BcryptHasher struct{}

// Hash returns a bcrypt hash of the password.
func (BcryptHasher) Hash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(h), err
}

// Verify checks a password against a bcrypt hash.
func (BcryptHasher) Verify(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
