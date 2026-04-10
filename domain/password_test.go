package domain

import (
	"errors"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "motdepasse1", false},
		{"valid with multiple digits", "abc12345", false},
		{"too short", "pass1", true},
		{"no digit", "motdepasse", true},
		{"only digits", "12345678", false},
		{"7 chars with digit", "abcde1f", true},
		{"8 chars no digit", "abcdefgh", true},
		{"8 chars with digit", "abcdefg1", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr && !errors.Is(err, ErrPasswordTooWeak) {
				t.Errorf("got %v, want ErrPasswordTooWeak", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}
