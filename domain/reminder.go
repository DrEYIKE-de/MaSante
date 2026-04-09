package domain

import (
	"context"
	"errors"
	"time"
)

// ReminderType identifies when a reminder is sent relative to the appointment.
type ReminderType string

const (
	ReminderJ7   ReminderType = "j7"
	ReminderJ2   ReminderType = "j2"
	ReminderJ0   ReminderType = "j0"
	ReminderLate ReminderType = "retard"
)

// ReminderStatus tracks the delivery state of a reminder.
type ReminderStatus string

const (
	ReminderScheduled ReminderStatus = "planifie"
	ReminderSent      ReminderStatus = "envoye"
	ReminderDelivered ReminderStatus = "recu"
	ReminderFailed    ReminderStatus = "echec"
)

// Reminder is a scheduled message to a patient about an upcoming appointment.
type Reminder struct {
	ID            int64
	AppointmentID int64
	PatientID     int64
	Channel       ReminderChannel
	Type          ReminderType
	Message       string
	Status        ReminderStatus
	ScheduledAt   time.Time
	SentAt        *time.Time
	ProviderID    string
	ErrorMessage  string
	RetryCount    int
	CreatedAt     time.Time

	PatientName string // denormalised
}

// MessageTemplate is a reusable SMS/WhatsApp message with placeholders.
type MessageTemplate struct {
	ID        int64
	Name      string // rappel_j7, rappel_j2, rappel_j0, retard
	Channel   ReminderChannel
	Body      string // contains {prenom}, {date}, {heure}, {centre}
	Language  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReminderStats holds delivery metrics.
type ReminderStats struct {
	DeliveryRate float64
	ConfirmRate  float64
	PendingCount int
	FailedCount  int
}

// ReminderRepository is a driven port for reminder persistence.
type ReminderRepository interface {
	Create(ctx context.Context, r *Reminder) error
	GetByID(ctx context.Context, id int64) (*Reminder, error)
	Update(ctx context.Context, r *Reminder) error
	ListPending(ctx context.Context) ([]Reminder, error)
	ListByAppointment(ctx context.Context, appointmentID int64) ([]Reminder, error)
	ListByPatient(ctx context.Context, patientID int64) ([]Reminder, error)
	Stats(ctx context.Context) (ReminderStats, error)
	ListTemplates(ctx context.Context) ([]MessageTemplate, error)
	UpdateTemplate(ctx context.Context, t *MessageTemplate) error
}

// SMSProvider is a driven port for sending text messages.
type SMSProvider interface {
	Send(ctx context.Context, to string, message string) (providerID string, err error)
	Name() string
}

var (
	ErrReminderNotFound = errors.New("rappel introuvable")
	ErrSMSProviderDown  = errors.New("fournisseur SMS indisponible")
)
