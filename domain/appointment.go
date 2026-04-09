package domain

import (
	"context"
	"errors"
	"time"
)

// AppointmentStatus tracks where an appointment is in its lifecycle.
type AppointmentStatus string

const (
	StatusConfirmed AppointmentStatus = "confirme"
	StatusPending   AppointmentStatus = "en_attente"
	StatusCompleted AppointmentStatus = "termine"
	StatusMissed    AppointmentStatus = "manque"
	StatusCancelled AppointmentStatus = "annule"
	StatusPostponed AppointmentStatus = "reporte"
)

// AppointmentType categorises the reason for a visit.
type AppointmentType string

const (
	TypeConsultation AppointmentType = "consultation"
	TypeMedPickup    AppointmentType = "retrait_medicaments"
	TypeBloodTest    AppointmentType = "bilan_sanguin"
	TypeAdherence    AppointmentType = "club_adherence"
)

// FollowUpFreq is how often a stable patient should return.
type FollowUpFreq string

const (
	FreqMonthly   FollowUpFreq = "mensuel"
	FreqQuarterly FollowUpFreq = "trimestriel"
	FreqBiannual  FollowUpFreq = "semestriel"
)

// Appointment represents a scheduled visit.
type Appointment struct {
	ID           int64
	PatientID    int64
	UserID       *int64 // assigned clinician
	Date         time.Time
	Time         string // HH:MM
	Type         AppointmentType
	Status       AppointmentStatus
	Notes        string
	FollowUpFreq *FollowUpFreq
	CreatedBy    *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Denormalised, populated by joins.
	PatientName string
	PatientCode string
}

// CompleteRequest holds data for marking an appointment as done.
type CompleteRequest struct {
	Notes        string
	FollowUpFreq FollowUpFreq
	NextDate     *time.Time
	NextType     AppointmentType
}

// MissedRequest holds follow-up actions when a patient doesn't show up.
type MissedRequest struct {
	SendReminder   bool
	AssignASC      bool
	Reschedule     bool
	RescheduleDays int
	Notes          string
}

// RescheduleRequest moves an appointment to a new date.
type RescheduleRequest struct {
	NewDate time.Time
	Reason  string
}

// AppointmentFilter holds criteria for listing appointments.
type AppointmentFilter struct {
	PatientID *int64
	DateFrom  *time.Time
	DateTo    *time.Time
	Status    *AppointmentStatus
	Page      int
	PerPage   int
}

// Slot represents a bookable time slot on a given day.
type Slot struct {
	Time      string
	Available bool
}

// AppointmentRepository is a driven port for appointment persistence.
type AppointmentRepository interface {
	Create(ctx context.Context, a *Appointment) error
	GetByID(ctx context.Context, id int64) (*Appointment, error)
	Update(ctx context.Context, a *Appointment) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f AppointmentFilter) ([]Appointment, int, error)
	ListByDate(ctx context.Context, date time.Time) ([]Appointment, error)
	ListByWeek(ctx context.Context, start time.Time) ([]Appointment, error)
	ListOverdue(ctx context.Context) ([]Appointment, error)
	AvailableSlots(ctx context.Context, date time.Time) ([]Slot, error)
	CountTodayByStatus(ctx context.Context) (map[AppointmentStatus]int, error)
}

var (
	ErrAppointmentNotFound = errors.New("rendez-vous introuvable")
	ErrSlotUnavailable     = errors.New("creneau indisponible")
)
