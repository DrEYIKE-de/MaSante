package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/masante/masante/domain"
)

// ReminderService orchestrates reminder generation, sending, and retry.
type ReminderService struct {
	reminders    domain.ReminderRepository
	appointments domain.AppointmentRepository
	patients     domain.PatientRepository
	smsConfig    domain.SMSConfigRepository
	center       domain.CenterRepository

	mu       sync.RWMutex
	provider domain.SMSProvider
}

// NewReminderService returns a new ReminderService.
func NewReminderService(
	reminders domain.ReminderRepository,
	appointments domain.AppointmentRepository,
	patients domain.PatientRepository,
	smsConfig domain.SMSConfigRepository,
	center domain.CenterRepository,
) *ReminderService {
	return &ReminderService{
		reminders:    reminders,
		appointments: appointments,
		patients:     patients,
		smsConfig:    smsConfig,
		center:       center,
	}
}

// SetProvider injects the SMS provider at runtime (after setup or config reload).
func (s *ReminderService) SetProvider(p domain.SMSProvider) {
	s.mu.Lock()
	s.provider = p
	s.mu.Unlock()
}

func (s *ReminderService) getProvider() domain.SMSProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider
}

// GenerateReminders creates reminder records for upcoming appointments
// that don't already have one. Called periodically by the scheduler.
func (s *ReminderService) GenerateReminders(ctx context.Context) error {
	cfg, err := s.smsConfig.Get(ctx)
	if err != nil || !cfg.Enabled {
		return err
	}

	templates, err := s.reminders.ListTemplates(ctx)
	if err != nil {
		return fmt.Errorf("list templates: %w", err)
	}
	tplMap := make(map[string]string)
	for _, t := range templates {
		if t.IsActive {
			tplMap[t.Name] = t.Body
		}
	}

	center, err := s.center.Get(ctx)
	if err != nil {
		return fmt.Errorf("get center: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	checks := []struct {
		rtype   domain.ReminderType
		tplName string
		enabled bool
		daysBefore int
	}{
		{domain.ReminderJ7, "rappel_j7", cfg.ReminderJ7, 7},
		{domain.ReminderJ2, "rappel_j2", cfg.ReminderJ2, 2},
		{domain.ReminderJ0, "rappel_j0", cfg.ReminderJ0, 0},
	}

	for _, check := range checks {
		if !check.enabled {
			continue
		}
		tpl, ok := tplMap[check.tplName]
		if !ok {
			continue
		}

		targetDate := today.AddDate(0, 0, check.daysBefore)
		apts, err := s.appointments.ListByDate(ctx, targetDate)
		if err != nil {
			return fmt.Errorf("list appointments for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		for _, apt := range apts {
			if apt.Status == domain.StatusCancelled || apt.Status == domain.StatusCompleted {
				continue
			}

			existing, _ := s.reminders.ListByAppointment(ctx, apt.ID)
			alreadySent := false
			for _, r := range existing {
				if r.Type == check.rtype {
					alreadySent = true
					break
				}
			}
			if alreadySent {
				continue
			}

			patient, err := s.patients.GetByID(ctx, apt.PatientID)
			if err != nil || patient.ReminderChannel == domain.ChannelNone || patient.Phone == "" {
				continue
			}

			msg := renderTemplate(tpl, patient, apt, center)

			reminder := &domain.Reminder{
				AppointmentID: apt.ID,
				PatientID:     patient.ID,
				Channel:       patient.ReminderChannel,
				Type:          check.rtype,
				Message:       msg,
				Status:        domain.ReminderScheduled,
				ScheduledAt:   time.Now(),
			}
			if err := s.reminders.Create(ctx, reminder); err != nil {
				return fmt.Errorf("create reminder: %w", err)
			}
		}
	}

	return nil
}

// ProcessQueue sends all pending reminders whose scheduled time has passed.
func (s *ReminderService) ProcessQueue(ctx context.Context) error {
	p := s.getProvider()
	if p == nil {
		return nil
	}

	pending, err := s.reminders.ListPending(ctx)
	if err != nil {
		return fmt.Errorf("list pending: %w", err)
	}

	for i := range pending {
		r := &pending[i]
		if time.Now().Before(r.ScheduledAt) {
			continue
		}

		patient, err := s.patients.GetByID(ctx, r.PatientID)
		if err != nil || patient.Phone == "" {
			r.Status = domain.ReminderFailed
			r.ErrorMessage = "patient introuvable ou sans telephone"
			s.reminders.Update(ctx, r)
			continue
		}

		providerID, err := p.Send(ctx, patient.Phone, r.Message)
		now := time.Now()
		r.SentAt = &now

		if err != nil {
			r.Status = domain.ReminderFailed
			r.ErrorMessage = err.Error()
			r.RetryCount++
		} else {
			r.Status = domain.ReminderSent
			r.ProviderID = providerID
		}

		s.reminders.Update(ctx, r)
	}

	return nil
}

// RetryFailed retries failed reminders up to 3 attempts with backoff.
func (s *ReminderService) RetryFailed(ctx context.Context) error {
	p := s.getProvider()
	if p == nil {
		return nil
	}

	pending, err := s.reminders.ListPending(ctx)
	if err != nil {
		return err
	}

	for i := range pending {
		r := &pending[i]
		if r.Status != domain.ReminderFailed || r.RetryCount >= 3 {
			continue
		}

		patient, err := s.patients.GetByID(ctx, r.PatientID)
		if err != nil || patient.Phone == "" {
			continue
		}

		providerID, err := p.Send(ctx, patient.Phone, r.Message)
		now := time.Now()
		r.SentAt = &now
		r.RetryCount++

		if err != nil {
			r.ErrorMessage = err.Error()
		} else {
			r.Status = domain.ReminderSent
			r.ProviderID = providerID
			r.ErrorMessage = ""
		}

		s.reminders.Update(ctx, r)
	}

	return nil
}

// Stats returns delivery metrics.
func (s *ReminderService) Stats(ctx context.Context) (domain.ReminderStats, error) {
	return s.reminders.Stats(ctx)
}

// ListPending returns reminders waiting to be sent.
func (s *ReminderService) ListPending(ctx context.Context) ([]domain.Reminder, error) {
	return s.reminders.ListPending(ctx)
}

// ListTemplates returns all message templates.
func (s *ReminderService) ListTemplates(ctx context.Context) ([]domain.MessageTemplate, error) {
	return s.reminders.ListTemplates(ctx)
}

// UpdateTemplate modifies a message template.
func (s *ReminderService) UpdateTemplate(ctx context.Context, t *domain.MessageTemplate) error {
	return s.reminders.UpdateTemplate(ctx, t)
}

// SendTest sends a one-off test SMS to verify provider configuration.
func (s *ReminderService) SendTest(ctx context.Context, to, message string) error {
	p := s.getProvider()
	if p == nil {
		return domain.ErrSMSProviderDown
	}
	_, err := p.Send(ctx, to, message)
	return err
}

func renderTemplate(tpl string, p *domain.Patient, a domain.Appointment, c *domain.Center) string {
	centerName := ""
	if c != nil {
		centerName = c.Name
	}
	r := strings.NewReplacer(
		"{prenom}", p.FirstName,
		"{nom}", p.LastName,
		"{date}", a.Date.Format("02/01/2006"),
		"{heure}", a.Time,
		"{centre}", centerName,
	)
	return r.Replace(tpl)
}
