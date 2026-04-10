package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/masante/masante/domain"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Title    string `json:"title"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "identifiant et mot de passe requis")
		return
	}

	session, user, err := s.auth.Login(r.Context(), req.Username, req.Password, r.RemoteAddr, r.UserAgent())
	if err != nil {
		if errors.Is(err, domain.ErrAccountLocked) {
			writeError(w, http.StatusTooManyRequests, "compte verrouille temporairement — reessayez dans 15 minutes")
			return
		}
		writeError(w, http.StatusUnauthorized, "identifiant ou mot de passe incorrect")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "masante_session",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, meResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     string(user.Role),
		Title:    user.Title,
		Phone:    user.Phone,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("masante_session")
	if err == nil {
		_ = s.auth.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "masante_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	writeJSON(w, http.StatusOK, meResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     string(user.Role),
		Title:    user.Title,
		Phone:    user.Phone,
	})
}
