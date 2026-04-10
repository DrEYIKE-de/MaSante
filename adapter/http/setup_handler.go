package http

import (
	"errors"
	"net/http"

	"github.com/masante/masante/app"
	"github.com/masante/masante/domain"
)

func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	done, err := s.setup.IsSetupDone(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}
	step, _ := s.setup.GetSetupStep(r.Context())
	centerName := ""
	if center, err := s.setup.GetCenter(r.Context()); err == nil && center != nil {
		centerName = center.Name
	}
	writeJSON(w, http.StatusOK, map[string]any{"setup_complete": done, "current_step": step, "center_name": centerName})
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
		if errors.Is(err, app.ErrSetupWrongStep) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
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
	if err := domain.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.setup.CreateAdmin(r.Context(), req); err != nil {
		if errors.Is(err, app.ErrSetupWrongStep) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
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
		if errors.Is(err, app.ErrSetupWrongStep) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
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
		if errors.Is(err, app.ErrSetupWrongStep) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "erreur lors de l'enregistrement")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	if err := s.setup.Complete(r.Context()); err != nil {
		if errors.Is(err, app.ErrSetupWrongStep) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "erreur lors de la finalisation")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
