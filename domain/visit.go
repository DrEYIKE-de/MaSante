package domain

import (
	"context"
	"time"
)

// PatientFound describes the outcome of a community health worker visit.
type PatientFound string

const (
	FoundYes   PatientFound = "oui"
	FoundNo    PatientFound = "non"
	FoundMoved PatientFound = "demenage"
)

// ASCVisit records a field visit by a community health worker.
type ASCVisit struct {
	ID                int64
	PatientID         int64
	ASCUserID         int64
	VisitDate         time.Time
	PatientFound      PatientFound
	AbsenceReason     string
	Notes             string
	NextAppointmentID *int64
	CreatedAt         time.Time

	// Denormalised, populated by joins.
	PatientName     string
	PatientDistrict string
	DaysOverdue     int
	ASCName         string
}

// ASCVisitRepository is a driven port for community health worker visits.
type ASCVisitRepository interface {
	Create(ctx context.Context, v *ASCVisit) error
	GetByID(ctx context.Context, id int64) (*ASCVisit, error)
	ListByASC(ctx context.Context, ascUserID int64) ([]ASCVisit, error)
	ListByPatient(ctx context.Context, patientID int64) ([]ASCVisit, error)
	ListPendingVisits(ctx context.Context) ([]ASCVisit, error)
}

// AuditEntry records a user action for traceability.
type AuditEntry struct {
	ID         int64
	UserID     *int64
	Action     string // e.g. "patient.create", "appointment.update"
	EntityType string // e.g. "patient", "user"
	EntityID   *int64
	Details    string // JSON
	IPAddress  string
	CreatedAt  time.Time
}

// AuditRepository is a driven port for the audit trail.
type AuditRepository interface {
	Log(ctx context.Context, e *AuditEntry) error
	ListByUser(ctx context.Context, userID int64, limit int) ([]AuditEntry, error)
}
