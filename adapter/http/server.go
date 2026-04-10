// Package http provides the driving adapter — the HTTP API that
// external clients (browser, mobile) use to interact with MaSante.
package http

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/masante/masante/app"
	"github.com/masante/masante/domain"
	"github.com/masante/masante/web"
)

// Server is the HTTP driving adapter. It translates HTTP requests
// into application service calls.
type Server struct {
	mux            *http.ServeMux
	auth           *app.AuthService
	setup          *app.SetupService
	patientSvc     *app.PatientService
	appointmentSvc *app.AppointmentService
	userSvc        *app.UserService
	reminderSvc    *app.ReminderService
}

// NewServer creates a Server and registers all routes.
func NewServer(
	auth *app.AuthService,
	setup *app.SetupService,
	patientSvc *app.PatientService,
	appointmentSvc *app.AppointmentService,
	userSvc *app.UserService,
	reminderSvc *app.ReminderService,
) *Server {
	s := &Server{
		mux:            http.NewServeMux(),
		auth:           auth,
		setup:          setup,
		patientSvc:     patientSvc,
		appointmentSvc: appointmentSvc,
		userSvc:        userSvc,
		reminderSvc:    reminderSvc,
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Setup (public, blocked after completion).
	s.mux.HandleFunc("GET /api/v1/setup/status", s.handleSetupStatus)
	s.mux.HandleFunc("POST /api/v1/setup/center", s.guardSetup(s.handleSetupCenter))
	s.mux.HandleFunc("POST /api/v1/setup/admin", s.guardSetup(s.handleSetupAdmin))
	s.mux.HandleFunc("POST /api/v1/setup/schedule", s.guardSetup(s.handleSetupSchedule))
	s.mux.HandleFunc("POST /api/v1/setup/sms", s.guardSetup(s.handleSetupSMS))
	s.mux.HandleFunc("POST /api/v1/setup/complete", s.guardSetup(s.handleSetupComplete))

	// Auth (public).
	s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	s.mux.HandleFunc("GET /api/v1/auth/me", s.requireAuth(s.handleMe))

	// Dashboard (authenticated).
	s.mux.HandleFunc("GET /api/v1/dashboard/stats", s.requireAuth(s.handleDashboardStats))
	s.mux.HandleFunc("GET /api/v1/dashboard/today", s.requireAuth(s.handleDashboardToday))
	s.mux.HandleFunc("GET /api/v1/dashboard/overdue", s.requireAuth(s.handleDashboardOverdue))

	// Users (admin only).
	onlyAdmin := s.requireRole(domain.RoleAdmin)
	s.mux.HandleFunc("GET /api/v1/users", onlyAdmin(s.handleListUsers))
	s.mux.HandleFunc("POST /api/v1/users", onlyAdmin(s.handleCreateUser))
	s.mux.HandleFunc("PUT /api/v1/users/{id}", onlyAdmin(s.handleUpdateUser))
	s.mux.HandleFunc("DELETE /api/v1/users/{id}", onlyAdmin(s.handleDisableUser))
	s.mux.HandleFunc("PUT /api/v1/users/{id}/reset-password", onlyAdmin(s.handleResetPassword))

	// Profile (authenticated, self).
	s.mux.HandleFunc("GET /api/v1/profile", s.requireAuth(s.handleGetProfile))
	s.mux.HandleFunc("PUT /api/v1/profile", s.requireAuth(s.handleUpdateProfile))
	s.mux.HandleFunc("PUT /api/v1/profile/password", s.requireAuth(s.handleChangePassword))
	s.mux.HandleFunc("GET /api/v1/profile/activity", s.requireAuth(s.handleProfileActivity))

	// Patients (authenticated).

	s.mux.HandleFunc("GET /api/v1/patients", s.requireAuth(s.handleListPatients))
	s.mux.HandleFunc("GET /api/v1/patients/search", s.requireAuth(s.handleSearchPatients))
	s.mux.HandleFunc("GET /api/v1/patients/{id}", s.requireAuth(s.handleGetPatient))
	s.mux.HandleFunc("POST /api/v1/patients", s.requireAuth(s.handleCreatePatient))
	s.mux.HandleFunc("PUT /api/v1/patients/{id}", s.requireAuth(s.handleUpdatePatient))
	s.mux.HandleFunc("PUT /api/v1/patients/{id}/exit", s.requireAuth(s.handleExitPatient))

	// Appointments (authenticated).
	s.mux.HandleFunc("POST /api/v1/appointments", s.requireAuth(s.handleCreateAppointment))
	s.mux.HandleFunc("GET /api/v1/appointments/{id}", s.requireAuth(s.handleGetAppointment))
	s.mux.HandleFunc("PUT /api/v1/appointments/{id}/complete", s.requireAuth(s.handleCompleteAppointment))
	s.mux.HandleFunc("PUT /api/v1/appointments/{id}/missed", s.requireAuth(s.handleMissedAppointment))
	s.mux.HandleFunc("PUT /api/v1/appointments/{id}/reschedule", s.requireAuth(s.handleRescheduleAppointment))
	s.mux.HandleFunc("DELETE /api/v1/appointments/{id}", s.requireAuth(s.handleCancelAppointment))
	s.mux.HandleFunc("GET /api/v1/appointments/slots", s.requireAuth(s.handleAvailableSlots))

	// Calendar (authenticated).
	s.mux.HandleFunc("GET /api/v1/calendar/week", s.requireAuth(s.handleCalendarWeek))

	// Reminders (authenticated).
	s.mux.HandleFunc("GET /api/v1/reminders", s.requireAuth(s.handleReminderQueue))
	s.mux.HandleFunc("GET /api/v1/reminders/stats", s.requireAuth(s.handleReminderStats))
	s.mux.HandleFunc("GET /api/v1/reminders/templates", s.requireAuth(s.handleReminderTemplates))
	s.mux.HandleFunc("PUT /api/v1/reminders/templates/{id}", s.requireAuth(s.handleUpdateTemplate))
	s.mux.HandleFunc("POST /api/v1/reminders/test", s.requireAuth(s.handleSendTestSMS))
	s.mux.HandleFunc("POST /api/v1/reminders/send-all", s.requireAuth(s.handleSendAllReminders))

	// Exports (authenticated).
	s.mux.HandleFunc("GET /api/v1/export/patients/excel", s.requireAuth(s.handleExportPatientsExcel))
	s.mux.HandleFunc("GET /api/v1/export/patients/pdf", s.requireAuth(s.handleExportPatientsPDF))
	s.mux.HandleFunc("GET /api/v1/export/monthly/excel", s.requireAuth(s.handleExportMonthlyExcel))
	s.mux.HandleFunc("GET /api/v1/export/monthly/pdf", s.requireAuth(s.handleExportMonthlyPDF))

	// Frontend — serve embedded files, fallback to index.html for SPA.
	frontendFS, _ := fs.Sub(web.Files, ".")
	fileServer := http.FileServer(http.FS(frontendFS))
	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// API routes are handled above; this catches everything else.
		fileServer.ServeHTTP(w, r)
	})
}

// guardSetup blocks access if the setup wizard is already complete.
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
