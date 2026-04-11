package http

import (
	"net/http"

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
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
