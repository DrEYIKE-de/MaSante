package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/masante/masante/adapter/export"
	"github.com/masante/masante/domain"
)

func (s *Server) handleExportPatientsExcel(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	f := domain.PatientFilter{Page: 1, PerPage: 10000}
	if status != "" {
		st := domain.PatientStatus(status)
		f.Status = &st
	}

	patients, _, err := s.patientSvc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	title := "Patients"
	if status != "" {
		title += " — " + status
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=patients_%s.xlsx", time.Now().Format("2006-01-02")))
	if err := export.PatientsToExcel(w, patients, title); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) handleExportPatientsPDF(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	f := domain.PatientFilter{Page: 1, PerPage: 10000}
	if status != "" {
		st := domain.PatientStatus(status)
		f.Status = &st
	}

	patients, _, err := s.patientSvc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	title := "Patients"
	if status != "" {
		title += " — " + status
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=patients_%s.pdf", time.Now().Format("2006-01-02")))
	if err := export.PatientsToPDF(w, patients, title); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) handleExportMonthlyExcel(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().AddDate(0, -1, 0).Format("2006-01")
	}

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

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=rapport_%s.xlsx", month))
	if err := export.MonthlyReportExcel(w, month, patientCounts, aptCounts); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) handleExportMonthlyPDF(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().AddDate(0, -1, 0).Format("2006-01")
	}

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

	centerName := "MaSante"
	// Could fetch from center repo if injected, for now keep simple.

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=rapport_%s.pdf", month))
	if err := export.MonthlyReportPDF(w, month, centerName, patientCounts, aptCounts); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
