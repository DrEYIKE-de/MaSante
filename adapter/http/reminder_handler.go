package http

import (
	"fmt"
	"net/http"
)

type testSMSRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

type updateTemplateRequest struct {
	Body     string `json:"body"`
	IsActive bool   `json:"is_active"`
}

func (s *Server) handleReminderQueue(w http.ResponseWriter, r *http.Request) {
	pending, err := s.reminderSvc.ListPending(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pending)
}

func (s *Server) handleReminderStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.reminderSvc.Stats(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleReminderTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := s.reminderSvc.ListTemplates(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

func (s *Server) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	var req updateTemplateRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	templates, err := s.reminderSvc.ListTemplates(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	id := r.PathValue("id")
	for _, t := range templates {
		if t.Name == id || fmt.Sprintf("%d", t.ID) == id {
			t.Body = req.Body
			t.IsActive = req.IsActive
			if err := s.reminderSvc.UpdateTemplate(r.Context(), &t); err != nil {
				internalError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}
	}

	writeError(w, http.StatusNotFound, "modele introuvable")
}

func (s *Server) handleSendTestSMS(w http.ResponseWriter, r *http.Request) {
	var req testSMSRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if req.To == "" {
		writeError(w, http.StatusBadRequest, "numero requis")
		return
	}

	msg := req.Message
	if msg == "" {
		msg = "Ceci est un message de test de MaSante. Si vous recevez ce message, la configuration SMS fonctionne."
	}

	if err := s.reminderSvc.SendTest(r.Context(), req.To, msg); err != nil {
		writeError(w, http.StatusServiceUnavailable, "fournisseur SMS indisponible")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSendAllReminders(w http.ResponseWriter, r *http.Request) {
	// Generate any pending reminders first, then send.
	if err := s.reminderSvc.GenerateReminders(r.Context()); err != nil {
		internalError(w, err)
		return
	}
	if err := s.reminderSvc.ProcessQueue(r.Context()); err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
