package http

import (
	"log"
	"net/http"

	"github.com/masante/masante/adapter/sms"
	"github.com/masante/masante/domain"
)

func (s *Server) handleGetSMSConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.setup.GetSMSConfig(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	// Don't expose API key/secret in full — mask them.
	masked := map[string]any{
		"enabled":   cfg.Enabled,
		"provider":  cfg.Provider,
		"sender_id": cfg.SenderID,
		"has_key":   cfg.APIKey != "",
		"j7":        cfg.ReminderJ7,
		"j2":        cfg.ReminderJ2,
		"j0":        cfg.ReminderJ0,
		"late":      cfg.ReminderLate,
	}
	writeJSON(w, http.StatusOK, masked)
}

func (s *Server) handleSaveSMSConfig(w http.ResponseWriter, r *http.Request) {
	var req domain.SetupSMSRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if err := s.setup.SaveSMSConfig(r.Context(), req); err != nil {
		internalError(w, err)
		return
	}

	// Reload SMS provider into the reminder service.
	cfg, err := s.setup.GetSMSConfig(r.Context())
	if err == nil && cfg.Enabled && cfg.Provider != "" {
		provider, err := sms.NewProvider(*cfg)
		if err != nil {
			log.Printf("[masante] SMS provider reload failed: %v", err)
		} else {
			s.reminderSvc.SetProvider(provider)
			log.Printf("[masante] SMS provider reloaded: %s", provider.Name())
		}
	} else if err == nil && !cfg.Enabled {
		s.reminderSvc.SetProvider(nil)
		log.Printf("[masante] SMS provider disabled")
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
