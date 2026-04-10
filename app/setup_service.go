package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/masante/masante/domain"
)

// ErrSetupWrongStep is returned when a setup endpoint is called out of order.
var ErrSetupWrongStep = errors.New("etape de configuration incorrecte")

// SetupService handles the first-launch configuration wizard.
// Steps must be called in order: 1=center, 2=admin, 3=schedule, 4=sms, 5=complete.
// Each step can only be called once and only at the right position.
type SetupService struct {
	center    domain.CenterRepository
	users     domain.UserRepository
	smsConfig domain.SMSConfigRepository
	hasher    domain.PasswordHasher
	audit     domain.AuditRepository
}

// NewSetupService returns a new SetupService.
func NewSetupService(
	center domain.CenterRepository,
	users domain.UserRepository,
	smsConfig domain.SMSConfigRepository,
	hasher domain.PasswordHasher,
	audit domain.AuditRepository,
) *SetupService {
	return &SetupService{
		center:    center,
		users:     users,
		smsConfig: smsConfig,
		hasher:    hasher,
		audit:     audit,
	}
}

// IsSetupDone reports whether the wizard has been completed.
func (s *SetupService) IsSetupDone(ctx context.Context) (bool, error) {
	return s.center.IsSetupDone(ctx)
}

// GetSetupStep returns the current step (0-5).
func (s *SetupService) GetSetupStep(ctx context.Context) (int, error) {
	return s.center.GetSetupStep(ctx)
}

// SaveCenter handles step 1. Creates the center record.
func (s *SetupService) SaveCenter(ctx context.Context, req domain.SetupCenterRequest) error {
	step, err := s.center.GetSetupStep(ctx)
	if err != nil {
		// No center row yet — step 0, that's correct for step 1.
		step = 0
	}
	if step != 0 {
		return fmt.Errorf("%w: attendu etape 1, actuellement a %d", ErrSetupWrongStep, step)
	}

	c := &domain.Center{
		Name:      req.Name,
		Type:      req.Type,
		Country:   req.Country,
		City:      req.City,
		District:  req.District,
		Latitude:  req.Lat,
		Longitude: req.Lng,
		SetupStep: 1,
	}
	if err := s.center.Create(ctx, c); err != nil {
		return err
	}
	return s.center.SetSetupStep(ctx, 1)
}

// CreateAdmin handles step 2. Creates the admin user account.
func (s *SetupService) CreateAdmin(ctx context.Context, req domain.SetupAdminRequest) error {
	if err := s.requireStep(ctx, 1); err != nil {
		return err
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     req.Username,
		PasswordHash: hash,
		FullName:     req.FullName,
		Email:        req.Email,
		Role:         domain.RoleAdmin,
		Title:        req.Title,
		Status:       domain.UserActive,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return err
	}
	return s.center.SetSetupStep(ctx, 2)
}

// SaveSchedule handles step 3. Configures consultation hours.
func (s *SetupService) SaveSchedule(ctx context.Context, req domain.SetupScheduleRequest) error {
	if err := s.requireStep(ctx, 2); err != nil {
		return err
	}

	c, err := s.center.Get(ctx)
	if err != nil {
		return fmt.Errorf("center not found: %w", err)
	}

	c.ConsultationDays = req.ConsultationDays
	c.StartTime = req.StartTime
	c.EndTime = req.EndTime
	c.SlotDuration = req.SlotDuration
	c.MaxPatientsDay = req.MaxPatientsDay
	c.UpdatedAt = time.Now()

	if err := s.center.Update(ctx, c); err != nil {
		return err
	}
	return s.center.SetSetupStep(ctx, 3)
}

// SaveSMSConfig handles step 4. Configures the SMS provider.
func (s *SetupService) SaveSMSConfig(ctx context.Context, req domain.SetupSMSRequest) error {
	if err := s.requireStep(ctx, 3); err != nil {
		return err
	}

	cfg := &domain.SMSConfig{
		Enabled:       req.Enabled,
		Provider:      req.Provider,
		APIKey:        req.APIKey,
		APISecret:     req.APISecret,
		SenderID:      req.SenderID,
		ReminderJ7:    true,
		ReminderJ2:    true,
		ReminderJ0:    false,
		ReminderLate:  true,
		LateDelayDays: 3,
	}
	if err := s.smsConfig.Save(ctx, cfg); err != nil {
		return err
	}
	return s.center.SetSetupStep(ctx, 4)
}

// Complete handles step 5. Finalizes the setup.
func (s *SetupService) Complete(ctx context.Context) error {
	if err := s.requireStep(ctx, 4); err != nil {
		return err
	}
	return s.center.CompleteSetup(ctx)
}

func (s *SetupService) requireStep(ctx context.Context, expected int) error {
	step, err := s.center.GetSetupStep(ctx)
	if err != nil {
		return fmt.Errorf("read setup step: %w", err)
	}
	if step != expected {
		return fmt.Errorf("%w: attendu etape %d, actuellement a %d", ErrSetupWrongStep, expected+1, step)
	}
	return nil
}
