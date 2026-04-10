package sms

import (
	"testing"

	"github.com/masante/masante/domain"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		provider string
		wantName string
		wantErr  bool
	}{
		{"africastalking", "africastalking", false},
		{"mtn", "mtn", false},
		{"orange", "orange", false},
		{"twilio", "twilio", false},
		{"infobip", "infobip", false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			cfg := domain.SMSConfig{
				Provider:  tt.provider,
				APIKey:    "test-key",
				APISecret: "test-secret",
				SenderID:  "MaSante",
			}

			p, err := NewProvider(cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != tt.wantName {
				t.Errorf("Name() = %q, want %q", p.Name(), tt.wantName)
			}
		})
	}
}
