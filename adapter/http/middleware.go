package http

import (
	"context"
	"net/http"

	"github.com/masante/masante/domain"
)

type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userContextKey).(*domain.User)
	return u
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("masante_session")
		if err != nil {
			writeError(w, http.StatusUnauthorized, "non authentifie")
			return
		}

		user, err := s.auth.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "session invalide")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) requireRole(roles ...domain.Role) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[domain.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil || !allowed[user.Role] {
				writeError(w, http.StatusForbidden, "acces refuse")
				return
			}
			next(w, r)
		})
	}
}
