package domain

import (
	"errors"
	"unicode"
)

// ErrPasswordTooWeak is returned when a password doesn't meet the policy.
var ErrPasswordTooWeak = errors.New("le mot de passe doit contenir au moins 8 caracteres dont au moins un chiffre")

// ValidatePassword checks that a password meets the minimum policy:
// at least 8 characters and at least one digit.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooWeak
	}
	hasDigit := false
	for _, r := range password {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return ErrPasswordTooWeak
	}
	return nil
}
