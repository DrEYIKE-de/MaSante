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

// GetSMSConfig returns the SMS configuration.
func (s *SetupService) GetSMSConfig(ctx context.Context) (*domain.SMSConfig, error) {
	return s.smsConfig.Get(ctx)
}

// GetCenter returns the center configuration.
func (s *SetupService) GetCenter(ctx context.Context) (*domain.Center, error) {
	return s.center.Get(ctx)
}

// GetSetupStep returns the current step (0-5).
func (s *SetupService) GetSetupStep(ctx context.Context) (int, error) {
	return s.center.GetSetupStep(ctx)
}

// SaveCenter handles step 1. Creates the center record.
func (s *SetupService) SaveCenter(ctx context.Context, req domain.SetupCenterRequest) error {
	step, _ := s.center.GetSetupStep(ctx)
	if step > 1 {
		return fmt.Errorf("%w: etape 1 deja validee", ErrSetupWrongStep)
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
	// Use Create or Update depending on whether center exists.
	existing, _ := s.center.Get(ctx)
	if existing != nil {
		c.ID = existing.ID
		if err := s.center.Update(ctx, c); err != nil {
			return err
		}
	} else {
		if err := s.center.Create(ctx, c); err != nil {
			return err
		}
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

	// Find any existing admin user (re-submission of step 2).
	// Search by role, not username, in case the username was changed.
	allUsers, _ := s.users.List(ctx)
	var existing *domain.User
	for i := range allUsers {
		if allUsers[i].Role == domain.RoleAdmin {
			existing = &allUsers[i]
			break
		}
	}

	if existing != nil {
		existing.Username = req.Username
		existing.PasswordHash = hash
		existing.FullName = req.FullName
		existing.Email = req.Email
		existing.Title = req.Title
		if err := s.users.Update(ctx, existing); err != nil {
			return err
		}
	} else {
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

func (s *SetupService) requireStep(ctx context.Context, minStep int) error {
	step, err := s.center.GetSetupStep(ctx)
	if err != nil {
		return fmt.Errorf("read setup step: %w", err)
	}
	if step < minStep {
		return fmt.Errorf("%w: completez d'abord l'etape %d", ErrSetupWrongStep, minStep)
	}
	return nil
}
