package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/masante/masante/app"
)

type Server struct {
	mux   *http.ServeMux
	auth  *app.AuthService
	setup *app.SetupService
}

func NewServer(auth *app.AuthService, setup *app.SetupService) *Server {
	s := &Server{
		mux:   http.NewServeMux(),
		auth:  auth,
		setup: setup,
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Setup (public, bloque si deja fait)
	s.mux.HandleFunc("GET /api/v1/setup/status", s.handleSetupStatus)
	s.mux.HandleFunc("POST /api/v1/setup/center", s.guardSetup(s.handleSetupCenter))
	s.mux.HandleFunc("POST /api/v1/setup/admin", s.guardSetup(s.handleSetupAdmin))
	s.mux.HandleFunc("POST /api/v1/setup/schedule", s.guardSetup(s.handleSetupSchedule))
	s.mux.HandleFunc("POST /api/v1/setup/sms", s.guardSetup(s.handleSetupSMS))
	s.mux.HandleFunc("POST /api/v1/setup/complete", s.guardSetup(s.handleSetupComplete))

	// Auth (public)
	s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	s.mux.HandleFunc("GET /api/v1/auth/me", s.requireAuth(s.handleMe))
}

// guardSetup bloque l'acces si le setup est deja fait.
func (s *Server) guardSetup(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		done, err := s.setup.IsSetupDone(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erreur interne")
			return
		}
		if done {
			writeError(w, http.StatusForbidden, "configuration deja effectuee")
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
