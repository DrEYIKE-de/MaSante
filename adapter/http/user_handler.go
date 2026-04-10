package http

import (
	"net/http"
	"strconv"

	"github.com/masante/masante/domain"
)

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Title    string `json:"title"`
}

type updateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Title    string `json:"title"`
	Status   string `json:"status"`
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.userSvc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur de lecture")
		return
	}

	// Strip password hashes from response.
	type safeUser struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
		Title    string `json:"title"`
		Status   string `json:"status"`
	}
	safe := make([]safeUser, len(users))
	for i, u := range users {
		safe[i] = safeUser{u.ID, u.Username, u.FullName, u.Email, u.Phone, string(u.Role), u.Title, string(u.Status)}
	}

	writeJSON(w, http.StatusOK, safe)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if req.Username == "" || req.Password == "" || req.FullName == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "username, password, full_name et role requis")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "mot de passe trop court (8 caracteres minimum)")
		return
	}

	u := &domain.User{
		Username: req.Username,
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Role:     domain.Role(req.Role),
		Title:    req.Title,
		Status:   domain.UserActive,
	}

	admin := UserFromContext(r.Context())
	if err := s.userSvc.Create(r.Context(), u, req.Password, admin.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la creation")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": u.ID, "username": u.Username})
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	u, err := s.userSvc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "utilisateur introuvable")
		return
	}

	var req updateUserRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	if req.FullName != "" {
		u.FullName = req.FullName
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	if req.Role != "" {
		u.Role = domain.Role(req.Role)
	}
	if req.Title != "" {
		u.Title = req.Title
	}
	if req.Status != "" {
		u.Status = domain.UserStatus(req.Status)
	}

	admin := UserFromContext(r.Context())
	if err := s.userSvc.Update(r.Context(), u, admin.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la mise a jour")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDisableUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	admin := UserFromContext(r.Context())
	if admin.ID == id {
		writeError(w, http.StatusBadRequest, "impossible de desactiver votre propre compte")
		return
	}

	if err := s.userSvc.Disable(r.Context(), id, admin.ID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	var req resetPasswordRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "mot de passe trop court")
		return
	}

	admin := UserFromContext(r.Context())
	if err := s.userSvc.ResetPassword(r.Context(), id, req.Password, admin.ID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
