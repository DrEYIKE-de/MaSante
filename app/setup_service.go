package app

import (
	"context"
	"fmt"
	"time"

	"github.com/masante/masante/domain"
)

type SetupService struct {
	center    domain.CenterRepository
	users     domain.UserRepository
	smsConfig domain.SMSConfigRepository
	hasher    domain.PasswordHasher
	audit     domain.AuditRepository
}

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

func (s *SetupService) IsSetupDone(ctx context.Context) (bool, error) {
	return s.center.IsSetupDone(ctx)
}

func (s *SetupService) SaveCenter(ctx context.Context, req domain.SetupCenterRequest) error {
	c := &domain.Center{
		Name:     req.Name,
		Type:     req.Type,
		Country:  req.Country,
		City:     req.City,
		District: req.District,
		Latitude: req.Lat,
		Longitude: req.Lng,
	}
	return s.center.Create(ctx, c)
}

func (s *SetupService) SaveSchedule(ctx context.Context, req domain.SetupScheduleRequest) error {
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

	return s.center.Update(ctx, c)
}

func (s *SetupService) CreateAdmin(ctx context.Context, req domain.SetupAdminRequest) error {
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

	return s.users.Create(ctx, user)
}

func (s *SetupService) SaveSMSConfig(ctx context.Context, req domain.SetupSMSRequest) error {
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
	return s.smsConfig.Save(ctx, cfg)
}

func (s *SetupService) Complete(ctx context.Context) error {
	return s.center.CompleteSetup(ctx)
}
