package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

type mockPatientRepo struct {
	patients map[int64]*domain.Patient
	nextID   int64
	seq      int
}

func newMockPatientRepo() *mockPatientRepo {
	return &mockPatientRepo{patients: make(map[int64]*domain.Patient), nextID: 1}
}

func (m *mockPatientRepo) Create(_ context.Context, p *domain.Patient) error {
	p.ID = m.nextID
	m.nextID++
	m.patients[p.ID] = p
	return nil
}
func (m *mockPatientRepo) GetByID(_ context.Context, id int64) (*domain.Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, domain.ErrPatientNotFound
	}
	return p, nil
}
func (m *mockPatientRepo) GetByCode(_ context.Context, code string) (*domain.Patient, error) {
	for _, p := range m.patients {
		if p.Code == code {
			return p, nil
		}
	}
	return nil, domain.ErrPatientNotFound
}
func (m *mockPatientRepo) Update(_ context.Context, p *domain.Patient) error {
	m.patients[p.ID] = p
	return nil
}
func (m *mockPatientRepo) List(_ context.Context, f domain.PatientFilter) ([]domain.Patient, int, error) {
	var result []domain.Patient
	for _, p := range m.patients {
		if f.Status != nil && p.Status != *f.Status {
			continue
		}
		result = append(result, *p)
	}
	return result, len(result), nil
}
func (m *mockPatientRepo) Search(_ context.Context, _ string, _ int) ([]domain.Patient, error) {
	return nil, nil
}
func (m *mockPatientRepo) NextCode(_ context.Context) (string, error) {
	m.seq++
	return "MS-2026-" + padInt(m.seq), nil
}
func (m *mockPatientRepo) CountByStatus(_ context.Context) (map[domain.PatientStatus]int, error) {
	counts := make(map[domain.PatientStatus]int)
	for _, p := range m.patients {
		counts[p.Status]++
	}
	return counts, nil
}

func padInt(n int) string {
	s := ""
	for i := 0; i < 5; i++ {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func TestPatientService_Create(t *testing.T) {
	repo := newMockPatientRepo()
	audit := &mockAudit{}
	svc := NewPatientService(repo, audit)

	p := &domain.Patient{
		LastName:  "Essomba",
		FirstName: "Nathalie",
		Sex:       "F",
		Phone:     "+237600000000",
	}

	if err := svc.Create(context.Background(), p, 1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == 0 {
		t.Error("ID should be set")
	}
	if p.Code == "" {
		t.Error("Code should be generated")
	}
	if p.Status != domain.PatientActive {
		t.Errorf("Status = %q, want active", p.Status)
	}
	if p.RiskScore != 5 {
		t.Errorf("RiskScore = %d, want 5", p.RiskScore)
	}
	if len(audit.entries) != 1 {
		t.Errorf("audit entries = %d, want 1", len(audit.entries))
	}
}

func TestPatientService_Exit(t *testing.T) {
	repo := newMockPatientRepo()
	audit := &mockAudit{}
	svc := NewPatientService(repo, audit)

	p := &domain.Patient{LastName: "Test", FirstName: "Patient", Sex: "M"}
	svc.Create(context.Background(), p, 1)

	req := domain.ExitRequest{
		Reason: domain.ExitDeath,
		Date:   time.Now(),
		Notes:  "deces constate",
	}
	if err := svc.Exit(context.Background(), p.ID, req, 1); err != nil {
		t.Fatalf("Exit: %v", err)
	}

	got, _ := svc.GetByID(context.Background(), p.ID)
	if got.Status != domain.PatientExited {
		t.Errorf("Status = %q, want sorti", got.Status)
	}
	if got.ExitReason == nil || *got.ExitReason != domain.ExitDeath {
		t.Error("ExitReason should be deces")
	}
}

func TestPatientService_Exit_NotFound(t *testing.T) {
	repo := newMockPatientRepo()
	svc := NewPatientService(repo, &mockAudit{})

	err := svc.Exit(context.Background(), 999, domain.ExitRequest{Reason: domain.ExitDeath, Date: time.Now()}, 1)
	if !errors.Is(err, domain.ErrPatientNotFound) {
		t.Errorf("got %v, want ErrPatientNotFound", err)
	}
}
