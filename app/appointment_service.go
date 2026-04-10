package app

import (
	"context"
	"fmt"
	"time"

	"github.com/masante/masante/domain"
)

// AppointmentService orchestrates appointment-related use cases.
type AppointmentService struct {
	appointments domain.AppointmentRepository
	patients     domain.PatientRepository
	audit        domain.AuditRepository
}

// NewAppointmentService returns a new AppointmentService.
func NewAppointmentService(
	appointments domain.AppointmentRepository,
	patients domain.PatientRepository,
	audit domain.AuditRepository,
) *AppointmentService {
	return &AppointmentService{
		appointments: appointments,
		patients:     patients,
		audit:        audit,
	}
}

// Schedule creates a new appointment after verifying slot availability.
func (s *AppointmentService) Schedule(ctx context.Context, a *domain.Appointment, createdBy int64) error {
	// Verify patient exists.
	if _, err := s.patients.GetByID(ctx, a.PatientID); err != nil {
		return fmt.Errorf("patient: %w", err)
	}

	// Check slot availability.
	slots, err := s.appointments.AvailableSlots(ctx, a.Date)
	if err != nil {
		return fmt.Errorf("check slots: %w", err)
	}

	available := false
	for _, slot := range slots {
		if slot.Time == a.Time && slot.Available {
			available = true
			break
		}
	}
	if !available {
		return domain.ErrSlotUnavailable
	}

	a.Status = domain.StatusConfirmed
	a.CreatedBy = &createdBy

	if err := s.appointments.Create(ctx, a); err != nil {
		return fmt.Errorf("create appointment: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &createdBy,
		Action:     "appointment.create",
		EntityType: "appointment",
		EntityID:   &a.ID,
	})
	return nil
}

// Complete marks an appointment as done, optionally scheduling the next one.
func (s *AppointmentService) Complete(ctx context.Context, id int64, req domain.CompleteRequest, doneBy int64) (*domain.Appointment, error) {
	a, err := s.appointments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	a.Status = domain.StatusCompleted
	a.Notes = req.Notes
	freq := req.FollowUpFreq
	a.FollowUpFreq = &freq

	if err := s.appointments.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update appointment: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &doneBy,
		Action:     "appointment.complete",
		EntityType: "appointment",
		EntityID:   &a.ID,
	})

	// Schedule next appointment if requested.
	var next *domain.Appointment
	if req.NextDate != nil {
		next = &domain.Appointment{
			PatientID: a.PatientID,
			Date:      *req.NextDate,
			Time:      a.Time,
			Type:      req.NextType,
			Status:    domain.StatusConfirmed,
		}
		if err := s.appointments.Create(ctx, next); err != nil {
			return nil, fmt.Errorf("schedule next: %w", err)
		}
	}

	return next, nil
}

// MarkMissed records a missed appointment and triggers follow-up actions.
// It returns the rescheduled appointment if Reschedule was requested.
func (s *AppointmentService) MarkMissed(ctx context.Context, id int64, req domain.MissedRequest, markedBy int64) (*domain.Appointment, error) {
	a, err := s.appointments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	a.Status = domain.StatusMissed
	a.Notes = req.Notes

	if err := s.appointments.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update appointment: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &markedBy,
		Action:     "appointment.missed",
		EntityType: "appointment",
		EntityID:   &a.ID,
	})

	// Reschedule if requested.
	var next *domain.Appointment
	if req.Reschedule && req.RescheduleDays > 0 {
		newDate := time.Now().AddDate(0, 0, req.RescheduleDays)
		next = &domain.Appointment{
			PatientID: a.PatientID,
			Date:      newDate,
			Time:      a.Time,
			Type:      a.Type,
			Status:    domain.StatusPending,
		}
		if err := s.appointments.Create(ctx, next); err != nil {
			return nil, fmt.Errorf("reschedule: %w", err)
		}
	}

	return next, nil
}

// Reschedule moves an appointment to a new date.
func (s *AppointmentService) Reschedule(ctx context.Context, id int64, req domain.RescheduleRequest, reschBy int64) error {
	a, err := s.appointments.GetByID(ctx, id)
	if err != nil {
		return err
	}

	a.Status = domain.StatusPostponed
	a.Notes = req.Reason

	if err := s.appointments.Update(ctx, a); err != nil {
		return fmt.Errorf("postpone original: %w", err)
	}

	// Create the new appointment.
	rescheduled := &domain.Appointment{
		PatientID: a.PatientID,
		UserID:    a.UserID,
		Date:      req.NewDate,
		Time:      a.Time,
		Type:      a.Type,
		Status:    domain.StatusConfirmed,
	}
	if err := s.appointments.Create(ctx, rescheduled); err != nil {
		return fmt.Errorf("create rescheduled: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &reschBy,
		Action:     "appointment.reschedule",
		EntityType: "appointment",
		EntityID:   &a.ID,
	})
	return nil
}

// Cancel deletes an appointment.
func (s *AppointmentService) Cancel(ctx context.Context, id int64, cancelledBy int64) error {
	a, err := s.appointments.GetByID(ctx, id)
	if err != nil {
		return err
	}

	a.Status = domain.StatusCancelled
	if err := s.appointments.Update(ctx, a); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &cancelledBy,
		Action:     "appointment.cancel",
		EntityType: "appointment",
		EntityID:   &a.ID,
	})
	return nil
}

// GetByID returns an appointment by ID.
func (s *AppointmentService) GetByID(ctx context.Context, id int64) (*domain.Appointment, error) {
	return s.appointments.GetByID(ctx, id)
}

// ListByDate returns all appointments for a given day.
func (s *AppointmentService) ListByDate(ctx context.Context, date time.Time) ([]domain.Appointment, error) {
	return s.appointments.ListByDate(ctx, date)
}

// ListByWeek returns all appointments for a 7-day period.
func (s *AppointmentService) ListByWeek(ctx context.Context, start time.Time) ([]domain.Appointment, error) {
	return s.appointments.ListByWeek(ctx, start)
}

// ListOverdue returns past appointments that were never completed.
func (s *AppointmentService) ListOverdue(ctx context.Context) ([]domain.Appointment, error) {
	return s.appointments.ListOverdue(ctx)
}

// AvailableSlots returns time slots for a given date.
func (s *AppointmentService) AvailableSlots(ctx context.Context, date time.Time) ([]domain.Slot, error) {
	return s.appointments.AvailableSlots(ctx, date)
}

// CountTodayByStatus returns today's appointment counts by status.
func (s *AppointmentService) CountTodayByStatus(ctx context.Context) (map[domain.AppointmentStatus]int, error) {
	return s.appointments.CountTodayByStatus(ctx)
}
