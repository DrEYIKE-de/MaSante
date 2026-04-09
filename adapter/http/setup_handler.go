package http

import (
	"net/http"

	"github.com/masante/masante/domain"
)

func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	done, err := s.setup.IsSetupDone(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"setup_complete": done})
}

func (s *Server) handleSetupCenter(w http.ResponseWriter, r *http.Request) {
	var req domain.SetupCenterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if req.Name == "" || req.Country == "" || req.City == "" {
		writeError(w, http.StatusBadRequest, "nom, pays et ville requis")
		return
	}
	if err := s.setup.SaveCenter(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de l'enregistrement")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSetupAdmin(w http.ResponseWriter, r *http.Request) {
	var req domain.SetupAdminRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if req.Username == "" || req.Password == "" || req.FullName == "" {
		writeError(w, http.StatusBadRequest, "nom, identifiant et mot de passe requis")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "mot de passe trop court (8 caracteres minimum)")
		return
	}
	if err := s.setup.CreateAdmin(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la creation du compte")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSetupSchedule(w http.ResponseWriter, r *http.Request) {
	var req domain.SetupScheduleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if err := s.setup.SaveSchedule(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de l'enregistrement")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSetupSMS(w http.ResponseWriter, r *http.Request) {
	var req domain.SetupSMSRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if err := s.setup.SaveSMSConfig(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de l'enregistrement")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	if err := s.setup.Complete(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la finalisation")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
