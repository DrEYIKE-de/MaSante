package http

import (
	"net/http"

	"github.com/masante/masante/domain"
)

type updateProfileRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	writeJSON(w, http.StatusOK, meResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     string(user.Role),
		Title:    user.Title,
	})
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	user := UserFromContext(r.Context())
	if err := s.userSvc.UpdateProfile(r.Context(), user.ID, req.FullName, req.Email, req.Phone); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la mise a jour")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if err := domain.ValidatePassword(req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := UserFromContext(r.Context())
	if err := s.userSvc.ChangePassword(r.Context(), user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusUnauthorized, "mot de passe actuel incorrect")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleProfileActivity(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	entries, err := s.userSvc.Activity(r.Context(), user.ID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur de lecture")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
