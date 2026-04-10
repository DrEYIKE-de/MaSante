package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

type mockAppointmentRepo struct {
	apts   map[int64]*domain.Appointment
	nextID int64
}

func newMockAppointmentRepo() *mockAppointmentRepo {
	return &mockAppointmentRepo{apts: make(map[int64]*domain.Appointment), nextID: 1}
}

func (m *mockAppointmentRepo) Create(_ context.Context, a *domain.Appointment) error {
	a.ID = m.nextID
	m.nextID++
	m.apts[a.ID] = a
	return nil
}
func (m *mockAppointmentRepo) GetByID(_ context.Context, id int64) (*domain.Appointment, error) {
	a, ok := m.apts[id]
	if !ok {
		return nil, domain.ErrAppointmentNotFound
	}
	return a, nil
}
func (m *mockAppointmentRepo) Update(_ context.Context, a *domain.Appointment) error {
	m.apts[a.ID] = a
	return nil
}
func (m *mockAppointmentRepo) Delete(_ context.Context, id int64) error {
	delete(m.apts, id)
	return nil
}
func (m *mockAppointmentRepo) List(_ context.Context, _ domain.AppointmentFilter) ([]domain.Appointment, int, error) {
	return nil, 0, nil
}
func (m *mockAppointmentRepo) ListByDate(_ context.Context, _ time.Time) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockAppointmentRepo) ListByWeek(_ context.Context, _ time.Time) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockAppointmentRepo) ListOverdue(_ context.Context) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockAppointmentRepo) AvailableSlots(_ context.Context, _ time.Time) ([]domain.Slot, error) {
	return []domain.Slot{
		{Time: "08:00", Available: true},
		{Time: "08:30", Available: true},
		{Time: "09:00", Available: false},
	}, nil
}
func (m *mockAppointmentRepo) CountTodayByStatus(_ context.Context) (map[domain.AppointmentStatus]int, error) {
	return nil, nil
}

func setupAptService() (*AppointmentService, *mockAppointmentRepo, *mockPatientRepo) {
	aptRepo := newMockAppointmentRepo()
	patRepo := newMockPatientRepo()
	// Seed a patient.
	patRepo.patients[1] = &domain.Patient{ID: 1, Code: "MS-2026-00001", LastName: "Test", FirstName: "Patient", Sex: "M", Status: domain.PatientActive}
	audit := &mockAudit{}
	svc := NewAppointmentService(aptRepo, patRepo, audit)
	return svc, aptRepo, patRepo
}

func TestAppointmentService_Schedule_OK(t *testing.T) {
	svc, _, _ := setupAptService()

	a := &domain.Appointment{
		PatientID: 1,
		Date:      time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		Time:      "08:00",
		Type:      domain.TypeConsultation,
	}
	if err := svc.Schedule(context.Background(), a, 1); err != nil {
		t.Fatalf("Schedule: %v", err)
	}
	if a.ID == 0 {
		t.Error("ID should be set")
	}
	if a.Status != domain.StatusConfirmed {
		t.Errorf("Status = %q, want confirme", a.Status)
	}
}

func TestAppointmentService_Schedule_SlotUnavailable(t *testing.T) {
	svc, _, _ := setupAptService()

	a := &domain.Appointment{
		PatientID: 1,
		Date:      time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		Time:      "09:00", // unavailable in mock
		Type:      domain.TypeConsultation,
	}
	err := svc.Schedule(context.Background(), a, 1)
	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("got %v, want ErrSlotUnavailable", err)
	}
}

func TestAppointmentService_Schedule_PatientNotFound(t *testing.T) {
	svc, _, _ := setupAptService()

	a := &domain.Appointment{
		PatientID: 999,
		Date:      time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		Time:      "08:00",
		Type:      domain.TypeConsultation,
	}
	err := svc.Schedule(context.Background(), a, 1)
	if err == nil {
		t.Fatal("expected error for unknown patient")
	}
}

func TestAppointmentService_Complete_WithNext(t *testing.T) {
	svc, aptRepo, _ := setupAptService()
	ctx := context.Background()

	a := &domain.Appointment{PatientID: 1, Date: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), Time: "08:00", Type: domain.TypeConsultation}
	svc.Schedule(ctx, a, 1)

	nextDate := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	next, err := svc.Complete(ctx, a.ID, domain.CompleteRequest{
		Notes:        "RAS",
		FollowUpFreq: domain.FreqQuarterly,
		NextDate:     &nextDate,
		NextType:     domain.TypeConsultation,
	}, 1)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	got, _ := aptRepo.GetByID(ctx, a.ID)
	if got.Status != domain.StatusCompleted {
		t.Errorf("original Status = %q, want termine", got.Status)
	}

	if next == nil {
		t.Fatal("next appointment should be created")
	}
	if next.Date != nextDate {
		t.Errorf("next date = %v, want %v", next.Date, nextDate)
	}
}

func TestAppointmentService_MarkMissed_WithReschedule(t *testing.T) {
	svc, _, _ := setupAptService()
	ctx := context.Background()

	a := &domain.Appointment{PatientID: 1, Date: time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC), Time: "08:00", Type: domain.TypeConsultation}
	svc.Schedule(ctx, a, 1)

	next, err := svc.MarkMissed(ctx, a.ID, domain.MissedRequest{
		Reschedule:     true,
		RescheduleDays: 7,
		Notes:          "patient absent",
	}, 1)
	if err != nil {
		t.Fatalf("MarkMissed: %v", err)
	}
	if next == nil {
		t.Fatal("rescheduled appointment should be created")
	}
	if next.Status != domain.StatusPending {
		t.Errorf("next Status = %q, want en_attente", next.Status)
	}
}

func TestAppointmentService_Reschedule(t *testing.T) {
	svc, aptRepo, _ := setupAptService()
	ctx := context.Background()

	a := &domain.Appointment{PatientID: 1, Date: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), Time: "08:00", Type: domain.TypeConsultation}
	svc.Schedule(ctx, a, 1)

	newDate := time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)
	if err := svc.Reschedule(ctx, a.ID, domain.RescheduleRequest{NewDate: newDate, Reason: "patient en voyage"}, 1); err != nil {
		t.Fatalf("Reschedule: %v", err)
	}

	original, _ := aptRepo.GetByID(ctx, a.ID)
	if original.Status != domain.StatusPostponed {
		t.Errorf("original Status = %q, want reporte", original.Status)
	}

	// New appointment should exist (ID = 3: seed schedule + reschedule).
	if len(aptRepo.apts) != 2 {
		t.Errorf("total appointments = %d, want 2", len(aptRepo.apts))
	}
}

func TestAppointmentService_Cancel(t *testing.T) {
	svc, aptRepo, _ := setupAptService()
	ctx := context.Background()

	a := &domain.Appointment{PatientID: 1, Date: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), Time: "08:00", Type: domain.TypeConsultation}
	svc.Schedule(ctx, a, 1)

	if err := svc.Cancel(ctx, a.ID, 1); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	got, _ := aptRepo.GetByID(ctx, a.ID)
	if got.Status != domain.StatusCancelled {
		t.Errorf("Status = %q, want annule", got.Status)
	}
}

func TestAppointmentService_Complete_NotFound(t *testing.T) {
	svc, _, _ := setupAptService()

	_, err := svc.Complete(context.Background(), 999, domain.CompleteRequest{}, 1)
	if !errors.Is(err, domain.ErrAppointmentNotFound) {
		t.Errorf("got %v, want ErrAppointmentNotFound", err)
	}
}
