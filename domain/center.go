package domain

import (
	"context"
	"time"
)

// CenterType classifies the health facility.
type CenterType string

const (
	CenterHospital CenterType = "hopital_public"
	CenterClinic   CenterType = "centre_sante"
	CenterPrivate  CenterType = "clinique_privee"
)

// Center holds configuration for the health facility running MaSante.
// There is exactly one Center per installation.
type Center struct {
	ID               int64
	Name             string
	Type             CenterType
	Country          string
	City             string
	District         string
	Latitude         *float64
	Longitude        *float64
	ConsultationDays string // comma-separated ISO weekdays: "1,2,3,4,5"
	StartTime        string // HH:MM
	EndTime          string // HH:MM
	SlotDuration     int    // minutes
	MaxPatientsDay   int
	SetupComplete    bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SMSConfig stores credentials and preferences for the SMS provider.
type SMSConfig struct {
	Enabled       bool
	Provider      string // africastalking, twilio, orange, mtn, infobip
	APIKey        string
	APISecret     string
	SenderID      string
	ReminderJ7    bool
	ReminderJ2    bool
	ReminderJ0    bool
	ReminderLate  bool
	LateDelayDays int
	UpdatedAt     time.Time
}

// SetupCenterRequest is the input for step 1 of the setup wizard.
type SetupCenterRequest struct {
	Name     string     `json:"name"`
	Type     CenterType `json:"type"`
	Country  string     `json:"country"`
	City     string     `json:"city"`
	District string     `json:"district"`
	Lat      *float64   `json:"lat"`
	Lng      *float64   `json:"lng"`
}

// SetupScheduleRequest is the input for step 3 of the setup wizard.
type SetupScheduleRequest struct {
	ConsultationDays string `json:"consultation_days"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	SlotDuration     int    `json:"slot_duration"`
	MaxPatientsDay   int    `json:"max_patients_day"`
}

// SetupAdminRequest is the input for step 2 of the setup wizard.
type SetupAdminRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Title    string `json:"title"`
}

// SetupSMSRequest is the input for step 4 of the setup wizard.
type SetupSMSRequest struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	SenderID  string `json:"sender_id"`
}

// CenterRepository is a driven port for center persistence.
type CenterRepository interface {
	Get(ctx context.Context) (*Center, error)
	Create(ctx context.Context, c *Center) error
	Update(ctx context.Context, c *Center) error
	CompleteSetup(ctx context.Context) error
	IsSetupDone(ctx context.Context) (bool, error)
}

// SMSConfigRepository is a driven port for SMS configuration persistence.
type SMSConfigRepository interface {
	Get(ctx context.Context) (*SMSConfig, error)
	Save(ctx context.Context, c *SMSConfig) error
}

// SettingsRepository is a driven port for generic key-value settings.
type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	All(ctx context.Context) (map[string]string, error)
}
