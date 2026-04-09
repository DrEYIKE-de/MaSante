package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

func seedPatient(t *testing.T, db *DB) int64 {
	t.Helper()
	repo := NewPatientRepo(db)
	p := makePatient("MS-2026-99999")
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("seed patient: %v", err)
	}
	return p.ID
}

func seedCenter(t *testing.T, db *DB) {
	t.Helper()
	repo := NewCenterRepo(db)
	repo.Create(context.Background(), &domain.Center{
		Name: "Test", Type: domain.CenterClinic, Country: "Cameroun", City: "Douala",
	})
}

func makeAppointment(patientID int64, date string, timeSlot string) *domain.Appointment {
	d, _ := time.Parse("2006-01-02", date)
	return &domain.Appointment{
		PatientID: patientID,
		Date:      d,
		Time:      timeSlot,
		Type:      domain.TypeConsultation,
		Status:    domain.StatusConfirmed,
	}
}

func TestAppointmentRepo_CreateAndGetByID(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	a := makeAppointment(pid, "2026-04-15", "10:00")
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.ID == 0 {
		t.Fatal("ID not set")
	}

	got, err := repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Time != "10:00" {
		t.Errorf("Time = %q, want 10:00", got.Time)
	}
	if got.PatientName == "" {
		t.Error("PatientName should be populated from join")
	}
}

func TestAppointmentRepo_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewAppointmentRepo(db)

	_, err := repo.GetByID(context.Background(), 999)
	if !errors.Is(err, domain.ErrAppointmentNotFound) {
		t.Errorf("got %v, want ErrAppointmentNotFound", err)
	}
}

func TestAppointmentRepo_Update(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	a := makeAppointment(pid, "2026-04-15", "09:00")
	repo.Create(ctx, a)

	a.Status = domain.StatusCompleted
	a.Notes = "RAS"
	freq := domain.FreqQuarterly
	a.FollowUpFreq = &freq
	if err := repo.Update(ctx, a); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, a.ID)
	if got.Status != domain.StatusCompleted {
		t.Errorf("Status = %q, want termine", got.Status)
	}
	if got.FollowUpFreq == nil || *got.FollowUpFreq != domain.FreqQuarterly {
		t.Error("FollowUpFreq should be trimestriel")
	}
}

func TestAppointmentRepo_Delete(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	a := makeAppointment(pid, "2026-04-15", "08:00")
	repo.Create(ctx, a)

	if err := repo.Delete(ctx, a.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, a.ID)
	if !errors.Is(err, domain.ErrAppointmentNotFound) {
		t.Error("appointment should be deleted")
	}
}

func TestAppointmentRepo_ListByDate(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	repo.Create(ctx, makeAppointment(pid, "2026-04-15", "08:00"))
	repo.Create(ctx, makeAppointment(pid, "2026-04-15", "09:00"))
	repo.Create(ctx, makeAppointment(pid, "2026-04-16", "10:00"))

	date, _ := time.Parse("2006-01-02", "2026-04-15")
	apts, err := repo.ListByDate(ctx, date)
	if err != nil {
		t.Fatalf("ListByDate: %v", err)
	}
	if len(apts) != 2 {
		t.Errorf("len = %d, want 2", len(apts))
	}
}

func TestAppointmentRepo_ListByWeek(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	repo.Create(ctx, makeAppointment(pid, "2026-04-13", "08:00"))
	repo.Create(ctx, makeAppointment(pid, "2026-04-15", "09:00"))
	repo.Create(ctx, makeAppointment(pid, "2026-04-19", "10:00"))
	repo.Create(ctx, makeAppointment(pid, "2026-04-21", "10:00")) // outside week

	start, _ := time.Parse("2006-01-02", "2026-04-13")
	apts, err := repo.ListByWeek(ctx, start)
	if err != nil {
		t.Fatalf("ListByWeek: %v", err)
	}
	if len(apts) != 3 {
		t.Errorf("len = %d, want 3", len(apts))
	}
}

func TestAppointmentRepo_ListOverdue(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	// Past confirmed = overdue
	repo.Create(ctx, makeAppointment(pid, "2026-01-01", "08:00"))
	// Past completed = not overdue
	a2 := makeAppointment(pid, "2026-01-02", "09:00")
	a2.Status = domain.StatusCompleted
	repo.Create(ctx, a2)
	// Future confirmed = not overdue
	repo.Create(ctx, makeAppointment(pid, "2099-01-01", "10:00"))

	overdue, err := repo.ListOverdue(ctx)
	if err != nil {
		t.Fatalf("ListOverdue: %v", err)
	}
	if len(overdue) != 1 {
		t.Errorf("len = %d, want 1", len(overdue))
	}
}

func TestAppointmentRepo_AvailableSlots(t *testing.T) {
	db := testDB(t)
	seedCenter(t, db)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	// Book the 08:00 slot
	repo.Create(ctx, makeAppointment(pid, "2026-04-15", "08:00"))

	date, _ := time.Parse("2006-01-02", "2026-04-15")
	slots, err := repo.AvailableSlots(ctx, date)
	if err != nil {
		t.Fatalf("AvailableSlots: %v", err)
	}

	if len(slots) == 0 {
		t.Fatal("expected slots")
	}

	// 08:00 should be unavailable
	for _, s := range slots {
		if s.Time == "08:00" && s.Available {
			t.Error("08:00 should be unavailable")
		}
		if s.Time == "08:30" && !s.Available {
			t.Error("08:30 should be available")
		}
	}
}

func TestAppointmentRepo_CountTodayByStatus(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	today := time.Now().Format("2006-01-02")
	a1 := makeAppointment(pid, today, "08:00")
	repo.Create(ctx, a1)
	a2 := makeAppointment(pid, today, "09:00")
	a2.Status = domain.StatusMissed
	repo.Create(ctx, a2)

	counts, err := repo.CountTodayByStatus(ctx)
	if err != nil {
		t.Fatalf("CountTodayByStatus: %v", err)
	}
	if counts[domain.StatusConfirmed] != 1 {
		t.Errorf("confirmed = %d, want 1", counts[domain.StatusConfirmed])
	}
	if counts[domain.StatusMissed] != 1 {
		t.Errorf("missed = %d, want 1", counts[domain.StatusMissed])
	}
}
