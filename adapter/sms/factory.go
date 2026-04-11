package sms

import (
	"fmt"

	"github.com/masante/masante/domain"
)

// NewProvider creates the appropriate SMSProvider based on the provider name.
func NewProvider(cfg domain.SMSConfig) (domain.SMSProvider, error) {
	switch cfg.Provider {
	case "africastalking":
		return NewAfricasTalking(cfg.APIKey, cfg.APISecret, cfg.SenderID), nil
	case "mtn":
		return NewMTN(cfg.APIKey, cfg.APISecret, cfg.SenderID), nil
	case "orange":
		return NewOrange(cfg.APIKey, cfg.SenderID), nil
	case "twilio":
		return NewTwilio(cfg.APIKey, cfg.APISecret, cfg.SenderID), nil
	default:
		return nil, fmt.Errorf("fournisseur SMS inconnu: %q", cfg.Provider)
	}
}
