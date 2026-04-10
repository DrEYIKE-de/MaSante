// Package app provides application services that orchestrate domain logic.
// Services depend only on domain ports, never on concrete adapters.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/masante/masante/domain"
)

// PatientService orchestrates patient-related use cases.
type PatientService struct {
	patients domain.PatientRepository
	audit    domain.AuditRepository
}

// NewPatientService returns a new PatientService.
func NewPatientService(patients domain.PatientRepository, audit domain.AuditRepository) *PatientService {
	return &PatientService{patients: patients, audit: audit}
}

// Create registers a new patient with an auto-generated code.
func (s *PatientService) Create(ctx context.Context, p *domain.Patient, createdBy int64) error {
	code, err := s.patients.NextCode(ctx)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}
	p.Code = code
	p.Status = domain.PatientActive
	p.RiskScore = 5
	p.EnrollmentDate = time.Now()

	if err := s.patients.Create(ctx, p); err != nil {
		return fmt.Errorf("create patient: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &createdBy,
		Action:     "patient.create",
		EntityType: "patient",
		EntityID:   &p.ID,
		Details:    p.Code,
	})
	return nil
}

// GetByID returns a patient by ID.
func (s *PatientService) GetByID(ctx context.Context, id int64) (*domain.Patient, error) {
	return s.patients.GetByID(ctx, id)
}

// Update persists changes to an existing patient.
func (s *PatientService) Update(ctx context.Context, p *domain.Patient, updatedBy int64) error {
	p.UpdatedAt = time.Now()
	if err := s.patients.Update(ctx, p); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &updatedBy,
		Action:     "patient.update",
		EntityType: "patient",
		EntityID:   &p.ID,
	})
	return nil
}

// Exit removes a patient from the active program.
func (s *PatientService) Exit(ctx context.Context, id int64, req domain.ExitRequest, exitBy int64) error {
	p, err := s.patients.GetByID(ctx, id)
	if err != nil {
		return err
	}

	p.Status = domain.PatientExited
	p.ExitReason = &req.Reason
	p.ExitDate = &req.Date
	p.ExitNotes = req.Notes
	p.UpdatedAt = time.Now()

	if err := s.patients.Update(ctx, p); err != nil {
		return fmt.Errorf("update patient exit: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &exitBy,
		Action:     "patient.exit",
		EntityType: "patient",
		EntityID:   &p.ID,
		Details:    string(req.Reason),
	})
	return nil
}

// List returns paginated patients matching the filter.
func (s *PatientService) List(ctx context.Context, f domain.PatientFilter) ([]domain.Patient, int, error) {
	return s.patients.List(ctx, f)
}

// Search performs a free-text search.
func (s *PatientService) Search(ctx context.Context, query string, limit int) ([]domain.Patient, error) {
	return s.patients.Search(ctx, query, limit)
}

// CountByStatus returns counts grouped by patient status.
func (s *PatientService) CountByStatus(ctx context.Context) (map[domain.PatientStatus]int, error) {
	return s.patients.CountByStatus(ctx)
}
