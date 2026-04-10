package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/masante/masante/domain"
)

type createAppointmentRequest struct {
	PatientID int64  `json:"patient_id"`
	Date      string `json:"date"` // YYYY-MM-DD
	Time      string `json:"time"` // HH:MM
	Type      string `json:"type"`
	Notes     string `json:"notes"`
}

type completeRequest struct {
	Notes        string `json:"notes"`
	FollowUpFreq string `json:"follow_up_freq"`
	NextDate     string `json:"next_date"`
	NextType     string `json:"next_type"`
}

type missedRequest struct {
	SendReminder   bool   `json:"send_reminder"`
	AssignASC      bool   `json:"assign_asc"`
	Reschedule     bool   `json:"reschedule"`
	RescheduleDays int    `json:"reschedule_days"`
	Notes          string `json:"notes"`
}

type rescheduleRequest struct {
	NewDate string `json:"new_date"`
	Reason  string `json:"reason"`
}

func (s *Server) handleCreateAppointment(w http.ResponseWriter, r *http.Request) {
	var req createAppointmentRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if req.PatientID == 0 || req.Date == "" || req.Time == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "patient_id, date, time et type requis")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "format de date invalide (YYYY-MM-DD)")
		return
	}

	a := &domain.Appointment{
		PatientID: req.PatientID,
		Date:      date,
		Time:      req.Time,
		Type:      domain.AppointmentType(req.Type),
		Notes:     req.Notes,
	}

	user := UserFromContext(r.Context())
	if err := s.appointmentSvc.Schedule(r.Context(), a, user.ID); err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrSlotUnavailable {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleGetAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	a, err := s.appointmentSvc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "rendez-vous introuvable")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleCompleteAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	var req completeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	domReq := domain.CompleteRequest{
		Notes:        req.Notes,
		FollowUpFreq: domain.FollowUpFreq(req.FollowUpFreq),
		NextType:     domain.AppointmentType(req.NextType),
	}
	if req.NextDate != "" {
		if d, err := time.Parse("2006-01-02", req.NextDate); err == nil {
			domReq.NextDate = &d
		}
	}

	user := UserFromContext(r.Context())
	next, err := s.appointmentSvc.Complete(r.Context(), id, domReq, user.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":           "ok",
		"next_appointment": next,
	})
}

func (s *Server) handleMissedAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	var req missedRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	user := UserFromContext(r.Context())
	next, err := s.appointmentSvc.MarkMissed(r.Context(), id, domain.MissedRequest{
		SendReminder:   req.SendReminder,
		AssignASC:      req.AssignASC,
		Reschedule:     req.Reschedule,
		RescheduleDays: req.RescheduleDays,
		Notes:          req.Notes,
	}, user.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":              "ok",
		"rescheduled_appointment": next,
	})
}

func (s *Server) handleRescheduleAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	var req rescheduleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	newDate, err := time.Parse("2006-01-02", req.NewDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "format de date invalide")
		return
	}

	user := UserFromContext(r.Context())
	if err := s.appointmentSvc.Reschedule(r.Context(), id, domain.RescheduleRequest{
		NewDate: newDate,
		Reason:  req.Reason,
	}, user.ID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCancelAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	user := UserFromContext(r.Context())
	if err := s.appointmentSvc.Cancel(r.Context(), id, user.ID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAvailableSlots(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "parametre date requis")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "format de date invalide")
		return
	}

	slots, err := s.appointmentSvc.AvailableSlots(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, slots)
}

func (s *Server) handleCalendarWeek(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "parametre date requis")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "format de date invalide")
		return
	}

	apts, err := s.appointmentSvc.ListByWeek(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, apts)
}

func (s *Server) handleDashboardToday(w http.ResponseWriter, r *http.Request) {
	apts, err := s.appointmentSvc.ListByDate(r.Context(), time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, apts)
}

func (s *Server) handleDashboardOverdue(w http.ResponseWriter, r *http.Request) {
	apts, err := s.appointmentSvc.ListOverdue(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, apts)
}

func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	patientCounts, err := s.patientSvc.CountByStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	aptCounts, err := s.appointmentSvc.CountTodayByStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"patients":     patientCounts,
		"appointments": aptCounts,
	})
}
