package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/masante/masante/domain"
)

type createPatientRequest struct {
	LastName        string `json:"last_name"`
	FirstName       string `json:"first_name"`
	DateOfBirth     string `json:"date_of_birth"` // YYYY-MM-DD
	Sex             string `json:"sex"`
	Phone           string `json:"phone"`
	PhoneSecondary  string `json:"phone_secondary"`
	District        string `json:"district"`
	Address         string `json:"address"`
	Language        string `json:"language"`
	ReminderChannel string `json:"reminder_channel"`
	ContactName     string `json:"contact_name"`
	ContactPhone    string `json:"contact_phone"`
	ContactRelation string `json:"contact_relation"`
	ReferredBy      string `json:"referred_by"`
}

type exitPatientRequest struct {
	Reason string `json:"reason"`
	Date   string `json:"date"`
	Notes  string `json:"notes"`
}

func (s *Server) handleCreatePatient(w http.ResponseWriter, r *http.Request) {
	var req createPatientRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}
	if req.LastName == "" || req.FirstName == "" || req.Sex == "" {
		writeError(w, http.StatusBadRequest, "nom, prenom et sexe requis")
		return
	}
	if err := domain.ValidateSex(req.Sex); err != nil {
		writeError(w, http.StatusBadRequest, "sexe invalide (M ou F)")
		return
	}
	if req.ReminderChannel != "" {
		if err := domain.ValidateReminderChannel(domain.ReminderChannel(req.ReminderChannel)); err != nil {
			writeError(w, http.StatusBadRequest, "canal de rappel invalide")
			return
		}
	}

	p := &domain.Patient{
		LastName:        req.LastName,
		FirstName:       req.FirstName,
		Sex:             req.Sex,
		Phone:           req.Phone,
		PhoneSecondary:  req.PhoneSecondary,
		District:        req.District,
		Address:         req.Address,
		Language:        req.Language,
		ReminderChannel: domain.ReminderChannel(req.ReminderChannel),
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactRelation: req.ContactRelation,
		ReferredBy:      req.ReferredBy,
	}
	if req.DateOfBirth != "" {
		if dob, err := time.Parse("2006-01-02", req.DateOfBirth); err == nil {
			p.DateOfBirth = &dob
		}
	}

	user := UserFromContext(r.Context())
	if err := s.patientSvc.Create(r.Context(), p, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de l'inscription")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleGetPatient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	p, err := s.patientSvc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "patient introuvable")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleUpdatePatient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	p, err := s.patientSvc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "patient introuvable")
		return
	}

	var req createPatientRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	if req.LastName != "" {
		p.LastName = req.LastName
	}
	if req.FirstName != "" {
		p.FirstName = req.FirstName
	}
	if req.Phone != "" {
		p.Phone = req.Phone
	}
	if req.District != "" {
		p.District = req.District
	}
	if req.Language != "" {
		p.Language = req.Language
	}
	if req.ReminderChannel != "" {
		p.ReminderChannel = domain.ReminderChannel(req.ReminderChannel)
	}

	user := UserFromContext(r.Context())
	if err := s.patientSvc.Update(r.Context(), p, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la mise a jour")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleExitPatient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id invalide")
		return
	}

	var req exitPatientRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "requete invalide")
		return
	}

	if err := domain.ValidateExitReason(domain.ExitReason(req.Reason)); err != nil {
		writeError(w, http.StatusBadRequest, "motif de sortie invalide")
		return
	}
	date, _ := time.Parse("2006-01-02", req.Date)
	if date.IsZero() {
		date = time.Now()
	}

	user := UserFromContext(r.Context())
	if err := s.patientSvc.Exit(r.Context(), id, domain.ExitRequest{
		Reason: domain.ExitReason(req.Reason),
		Date:   date,
		Notes:  req.Notes,
	}, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "ressource introuvable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListPatients(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.PatientFilter{
		District: q.Get("district"),
		Query:    q.Get("q"),
	}
	if v := q.Get("status"); v != "" {
		st := domain.PatientStatus(v)
		f.Status = &st
	}
	f.Page, _ = strconv.Atoi(q.Get("page"))
	f.PerPage, _ = strconv.Atoi(q.Get("per_page"))

	patients, total, err := s.patientSvc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la recherche")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"patients": patients,
		"total":    total,
		"page":     f.Page,
		"per_page": f.PerPage,
	})
}

func (s *Server) handleSearchPatients(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "parametre q requis")
		return
	}

	patients, err := s.patientSvc.Search(r.Context(), query, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur de recherche")
		return
	}

	writeJSON(w, http.StatusOK, patients)
}
