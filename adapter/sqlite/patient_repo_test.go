package sqlite

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

func makePatient(code string) *domain.Patient {
	return &domain.Patient{
		Code:            code,
		LastName:        "Essomba",
		FirstName:       "Nathalie",
		Sex:             "F",
		Phone:           "+237600000000",
		District:        "Akwa",
		Language:        "fr",
		ReminderChannel: domain.ChannelSMS,
		Status:          domain.PatientActive,
		RiskScore:       3,
		EnrollmentDate:  time.Now(),
	}
}

func TestPatientRepo_CreateAndGetByID(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	p := makePatient("MS-2026-00001")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("ID not set")
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LastName != "Essomba" {
		t.Errorf("LastName = %q, want Essomba", got.LastName)
	}
	if got.District != "Akwa" {
		t.Errorf("District = %q, want Akwa", got.District)
	}
	if got.ReminderChannel != domain.ChannelSMS {
		t.Errorf("ReminderChannel = %q, want sms", got.ReminderChannel)
	}
}

func TestPatientRepo_GetByCode(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	p := makePatient("MS-2026-00099")
	repo.Create(ctx, p)

	got, err := repo.GetByCode(ctx, "MS-2026-00099")
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("ID = %d, want %d", got.ID, p.ID)
	}
}

func TestPatientRepo_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)

	_, err := repo.GetByID(context.Background(), 999)
	if !errors.Is(err, domain.ErrPatientNotFound) {
		t.Errorf("got %v, want ErrPatientNotFound", err)
	}
}

func TestPatientRepo_Update(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	p := makePatient("MS-2026-00002")
	repo.Create(ctx, p)

	p.District = "Bonapriso"
	p.RiskScore = 7
	p.Status = domain.PatientMonitored
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, p.ID)
	if got.District != "Bonapriso" {
		t.Errorf("District = %q, want Bonapriso", got.District)
	}
	if got.RiskScore != 7 {
		t.Errorf("RiskScore = %d, want 7", got.RiskScore)
	}
	if got.Status != domain.PatientMonitored {
		t.Errorf("Status = %q, want a_surveiller", got.Status)
	}
}

func TestPatientRepo_Update_Exit(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	p := makePatient("MS-2026-00003")
	repo.Create(ctx, p)

	reason := domain.ExitDeath
	now := time.Now()
	p.Status = domain.PatientExited
	p.ExitReason = &reason
	p.ExitDate = &now
	p.ExitNotes = "deces constate"
	repo.Update(ctx, p)

	got, _ := repo.GetByID(ctx, p.ID)
	if got.ExitReason == nil || *got.ExitReason != domain.ExitDeath {
		t.Errorf("ExitReason = %v, want deces", got.ExitReason)
	}
	if got.ExitDate == nil {
		t.Error("ExitDate should be set")
	}
}

func TestPatientRepo_List_Pagination(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		p := makePatient(fmt.Sprintf("MS-2026-%05d", i))
		p.FirstName = fmt.Sprintf("Patient%d", i)
		repo.Create(ctx, p)
	}

	patients, total, err := repo.List(ctx, domain.PatientFilter{Page: 1, PerPage: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(patients) != 2 {
		t.Errorf("len = %d, want 2", len(patients))
	}
}

func TestPatientRepo_List_FilterByStatus(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	active := makePatient("MS-2026-00010")
	active.Status = domain.PatientActive
	repo.Create(ctx, active)

	lost := makePatient("MS-2026-00011")
	lost.Status = domain.PatientLost
	repo.Create(ctx, lost)

	status := domain.PatientLost
	patients, total, err := repo.List(ctx, domain.PatientFilter{Status: &status, Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(patients) != 1 || patients[0].Code != "MS-2026-00011" {
		t.Error("expected only the lost patient")
	}
}

func TestPatientRepo_Search(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	p1 := makePatient("MS-2026-00020")
	p1.LastName = "Mbouda"
	p1.FirstName = "Thierry"
	repo.Create(ctx, p1)

	p2 := makePatient("MS-2026-00021")
	p2.LastName = "Essomba"
	p2.FirstName = "Nathalie"
	repo.Create(ctx, p2)

	results, err := repo.Search(ctx, "mbouda", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1", len(results))
	}
	if results[0].LastName != "Mbouda" {
		t.Errorf("got %q, want Mbouda", results[0].LastName)
	}
}

func TestPatientRepo_NextCode(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	code1, err := repo.NextCode(ctx)
	if err != nil {
		t.Fatalf("NextCode: %v", err)
	}
	year := time.Now().Year()
	expected := fmt.Sprintf("MS-%d-00001", year)
	if code1 != expected {
		t.Errorf("code = %q, want %q", code1, expected)
	}

	// Insert one, next should be 00002.
	p := makePatient(code1)
	repo.Create(ctx, p)

	code2, _ := repo.NextCode(ctx)
	expected2 := fmt.Sprintf("MS-%d-00002", year)
	if code2 != expected2 {
		t.Errorf("code = %q, want %q", code2, expected2)
	}
}

func TestPatientRepo_CountByStatus(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	for i, s := range []domain.PatientStatus{domain.PatientActive, domain.PatientActive, domain.PatientLost} {
		p := makePatient(fmt.Sprintf("MS-2026-001%02d", i))
		p.Status = s
		repo.Create(ctx, p)
	}

	counts, err := repo.CountByStatus(ctx)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts[domain.PatientActive] != 2 {
		t.Errorf("active = %d, want 2", counts[domain.PatientActive])
	}
	if counts[domain.PatientLost] != 1 {
		t.Errorf("lost = %d, want 1", counts[domain.PatientLost])
	}
}
