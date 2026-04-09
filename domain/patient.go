// Package domain defines the core business types and ports (interfaces)
// for the MaSante application. It has no external dependencies.
package domain

import (
	"context"
	"errors"
	"time"
)

// PatientStatus represents the current state of a patient in the program.
type PatientStatus string

const (
	PatientActive    PatientStatus = "active"
	PatientMonitored PatientStatus = "a_surveiller"
	PatientLost      PatientStatus = "perdu_de_vue"
	PatientExited    PatientStatus = "sorti"
)

// ExitReason describes why a patient left the program.
type ExitReason string

const (
	ExitDeath    ExitReason = "deces"
	ExitTransfer ExitReason = "transfert"
	ExitDropout  ExitReason = "abandon"
	ExitLost     ExitReason = "perdu_de_vue"
	ExitCured    ExitReason = "guerison"
)

// ReminderChannel is the preferred communication channel for a patient.
type ReminderChannel string

const (
	ChannelSMS      ReminderChannel = "sms"
	ChannelWhatsApp ReminderChannel = "whatsapp"
	ChannelVoice    ReminderChannel = "voice"
	ChannelNone     ReminderChannel = "none"
)

// Patient is the central entity of the system.
type Patient struct {
	ID              int64
	Code            string // unique, format MS-YYYY-NNNNN
	LastName        string
	FirstName       string
	DateOfBirth     *time.Time
	Sex             string // M or F
	Phone           string
	PhoneSecondary  string
	District        string
	Address         string
	Language        string // fr, en, duala, ewondo, bamileke
	ReminderChannel ReminderChannel
	ContactName     string // trusted contact
	ContactPhone    string
	ContactRelation string
	ReferredBy      string
	Status          PatientStatus
	RiskScore       int // 0-10
	EnrollmentDate  time.Time
	ExitReason      *ExitReason
	ExitDate        *time.Time
	ExitNotes       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PatientFilter holds criteria for listing patients.
type PatientFilter struct {
	Status  *PatientStatus
	District string
	Query    string // free-text search on name, code, phone
	Page     int
	PerPage  int
}

// ExitRequest captures the data needed to remove a patient from the program.
type ExitRequest struct {
	Reason ExitReason
	Date   time.Time
	Notes  string
}

// PatientRepository is a driven port for patient persistence.
type PatientRepository interface {
	Create(ctx context.Context, p *Patient) error
	GetByID(ctx context.Context, id int64) (*Patient, error)
	GetByCode(ctx context.Context, code string) (*Patient, error)
	Update(ctx context.Context, p *Patient) error
	List(ctx context.Context, f PatientFilter) ([]Patient, int, error)
	Search(ctx context.Context, query string, limit int) ([]Patient, error)
	NextCode(ctx context.Context) (string, error)
	CountByStatus(ctx context.Context) (map[PatientStatus]int, error)
}

var (
	ErrPatientNotFound  = errors.New("patient introuvable")
	ErrPatientCodeTaken = errors.New("code patient deja utilise")
)
