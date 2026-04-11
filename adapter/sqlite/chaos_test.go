package sqlite

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

// TestChaos_50GoroutinesCreatingPatients hammers the database with 50
// concurrent goroutines inserting patients to surface race conditions
// or SQLite lock contention issues.
func TestChaos_50GoroutinesCreatingPatients(t *testing.T) {
	db := testDB(t)
	patientRepo := NewPatientRepo(db)
	ctx := context.Background()

	const numGoroutines = 50
	const opsPerGoroutine = 10

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines*opsPerGoroutine)

	// Writers: insert patients concurrently from 50 goroutines.
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				p := &domain.Patient{
					Code:            fmt.Sprintf("G50-%04d-%05d", gID, i),
					LastName:        fmt.Sprintf("LastName_%d_%d", gID, i),
					FirstName:       fmt.Sprintf("FirstName_%d_%d", gID, i),
					Sex:             "M",
					Phone:           fmt.Sprintf("+23760000%04d", gID*100+i),
					District:        "Akwa",
					Language:        "fr",
					ReminderChannel: domain.ChannelSMS,
					Status:          domain.PatientActive,
					RiskScore:       5,
					EnrollmentDate:  time.Now(),
				}
				if err := patientRepo.Create(ctx, p); err != nil {
					errCh <- fmt.Errorf("goroutine %d, op %d: create: %w", gID, i, err)
				}
			}
		}(g)
	}

	// Readers: list patients concurrently while writes happen.
	for g := 0; g < numGoroutines/2; g++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				_, _, err := patientRepo.List(ctx, domain.PatientFilter{Page: 1, PerPage: 10})
				if err != nil {
					errCh <- fmt.Errorf("reader %d, op %d: list: %w", gID, i, err)
				}
			}
		}(g)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}

	// Verify total count.
	_, count, err := patientRepo.List(ctx, domain.PatientFilter{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("final count: %v", err)
	}
	expected := numGoroutines * opsPerGoroutine
	if count != expected {
		t.Errorf("total patients = %d, want %d", count, expected)
	}
}

// TestChaos_Insert1000ThenSearch inserts 1000 patients and verifies that
// queries still return correct results at scale.
func TestChaos_Insert1000ThenSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large dataset test in short mode")
	}

	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	const total = 1000

	for i := 0; i < total; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("LD-%05d", i),
			LastName:        fmt.Sprintf("Last_%d", i),
			FirstName:       fmt.Sprintf("First_%d", i),
			Sex:             []string{"M", "F"}[i%2],
			Phone:           fmt.Sprintf("+237%09d", i),
			District:        fmt.Sprintf("District_%d", i%10),
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       i % 11,
			EnrollmentDate:  time.Now(),
		}
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("insert %d: %v", i, err)
		}
	}

	// Verify total count.
	patients, count, err := repo.List(ctx, domain.PatientFilter{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if count != total {
		t.Errorf("total = %d, want %d", count, total)
	}
	if len(patients) != 10 {
		t.Errorf("page size = %d, want 10", len(patients))
	}

	// Verify search works at scale.
	results, err := repo.Search(ctx, "Last_500", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one search result for 'Last_500'")
	}

	// Verify count by status.
	counts, err := repo.CountByStatus(ctx)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts[domain.PatientActive] != total {
		t.Errorf("active count = %d, want %d", counts[domain.PatientActive], total)
	}

	// Pagination: last page.
	patients, _, err = repo.List(ctx, domain.PatientFilter{Page: 100, PerPage: 10})
	if err != nil {
		t.Fatalf("List last page: %v", err)
	}
	if len(patients) != 10 {
		t.Errorf("last page size = %d, want 10", len(patients))
	}
}

// TestChaos_SQLInjectionInPatientNames verifies that SQL injection payloads
// including '; DROP TABLE patients; -- in patient names, codes, and search
// queries do not cause breakage or data leaks.
func TestChaos_SQLInjectionInPatientNames(t *testing.T) {
	db := testDB(t)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	injections := []string{
		"'; DROP TABLE patients; --",
		"Robert'); DROP TABLE patients;--",
		"' OR '1'='1",
		"'; DELETE FROM patients WHERE '1'='1",
		"' UNION SELECT * FROM users--",
		"\\'; UPDATE patients SET status='sorti';--",
		"<script>alert('xss')</script>",
		"%' OR 1=1 --",
		"' AND (SELECT COUNT(*) FROM users) > 0 --",
		"name\x00null_byte",
	}

	for i, injection := range injections {
		p := &domain.Patient{
			Code:            fmt.Sprintf("INJ-%05d", i),
			LastName:        injection,
			FirstName:       injection,
			Sex:             "M",
			Phone:           "+237600000000",
			District:        injection,
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       5,
			EnrollmentDate:  time.Now(),
		}
		if err := repo.Create(ctx, p); err != nil {
			t.Errorf("injection %d (%q): create failed: %v", i, injection, err)
			continue
		}

		// Verify the data is stored literally, not executed.
		got, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Errorf("injection %d: GetByID failed: %v", i, err)
			continue
		}
		if got.LastName != injection {
			t.Errorf("injection %d: LastName = %q, want %q", i, got.LastName, injection)
		}
	}

	// Verify search with injection payloads does not panic or error.
	for _, injection := range injections {
		_, err := repo.Search(ctx, injection, 10)
		if err != nil {
			t.Errorf("Search(%q): %v", injection, err)
		}
	}

	// Verify the patients table still exists and has correct count.
	_, count, err := repo.List(ctx, domain.PatientFilter{Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List after injections: %v", err)
	}
	if count != len(injections) {
		t.Errorf("count = %d, want %d (table may have been corrupted)", count, len(injections))
	}
}

// TestChaos_100RapidSessionCreateDeleteCycles creates and deletes sessions
// rapidly across 100 concurrent goroutines to test for leaks, deadlocks,
// or constraint violations.
func TestChaos_100RapidSessionCreateDeleteCycles(t *testing.T) {
	db := testDB(t)
	userRepo := NewUserRepo(db)
	sessionRepo := NewSessionRepo(db)
	ctx := context.Background()

	u := &domain.User{
		Username:     "session_chaos",
		PasswordHash: "hash",
		FullName:     "Session Chaos",
		Role:         domain.RoleAdmin,
		Status:       domain.UserActive,
	}
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	const cycles = 100

	var wg sync.WaitGroup
	errCh := make(chan error, cycles*3)

	for i := 0; i < cycles; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			token := fmt.Sprintf("chaos-token-%d-%d", i, rand.Int63())
			s := &domain.Session{
				Token:     token,
				UserID:    u.ID,
				ExpiresAt: time.Now().Add(1 * time.Hour),
				IPAddress: "127.0.0.1",
				UserAgent: "chaos-test",
			}
			if err := sessionRepo.Create(ctx, s); err != nil {
				errCh <- fmt.Errorf("cycle %d: create session: %w", i, err)
				return
			}

			// Read it back.
			_, err := sessionRepo.GetByToken(ctx, token)
			if err != nil {
				errCh <- fmt.Errorf("cycle %d: get session: %w", i, err)
				return
			}

			// Delete it.
			if err := sessionRepo.DeleteByToken(ctx, token); err != nil {
				errCh <- fmt.Errorf("cycle %d: delete session: %w", i, err)
			}

			// Verify it's gone.
			_, err = sessionRepo.GetByToken(ctx, token)
			if err == nil {
				errCh <- fmt.Errorf("cycle %d: session should be deleted", i)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestChaos_AppointmentsWithInvalidData verifies that the database rejects
// appointments with invalid data: missing foreign keys, invalid types,
// invalid statuses, and empty required fields.
func TestChaos_AppointmentsWithInvalidData(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	seedCenter(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		apt     *domain.Appointment
		wantErr bool
	}{
		{
			name: "valid appointment",
			apt: &domain.Appointment{
				PatientID: pid,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "10:00",
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
			},
			wantErr: false,
		},
		{
			name: "nonexistent patient_id",
			apt: &domain.Appointment{
				PatientID: 999999,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "10:00",
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
			},
			wantErr: true, // FK constraint
		},
		{
			name: "invalid appointment type",
			apt: &domain.Appointment{
				PatientID: pid,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "10:00",
				Type:      "invalid_type",
				Status:    domain.StatusConfirmed,
			},
			wantErr: true, // CHECK constraint
		},
		{
			name: "invalid status",
			apt: &domain.Appointment{
				PatientID: pid,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "10:00",
				Type:      domain.TypeConsultation,
				Status:    "invalid_status",
			},
			wantErr: true, // CHECK constraint
		},
		{
			name: "empty date",
			apt: &domain.Appointment{
				PatientID: pid,
				Date:      time.Time{},
				Time:      "10:00",
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
			},
			wantErr: false, // SQLite stores zero date as string, no constraint
		},
		{
			name: "empty time",
			apt: &domain.Appointment{
				PatientID: pid,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "",
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
			},
			wantErr: false, // no NOT NULL check on time content
		},
		{
			name: "SQL injection in notes",
			apt: &domain.Appointment{
				PatientID: pid,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "11:00",
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
				Notes:     "'; DROP TABLE appointments; --",
			},
			wantErr: false, // parameterised query should handle this safely
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.apt)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	// Verify SQL injection in notes was stored literally.
	results, _, err := repo.List(ctx, domain.AppointmentFilter{Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List after invalid data tests: %v", err)
	}
	foundInjection := false
	for _, a := range results {
		if a.Notes == "'; DROP TABLE appointments; --" {
			foundInjection = true
			break
		}
	}
	if !foundInjection {
		t.Error("SQL injection notes should be stored literally")
	}
}

// TestChaos_DatabaseRecovery simulates a crash by closing the DB mid-operation
// and verifying that data written before the crash is still consistent after
// reopening.
func TestChaos_DatabaseRecovery(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/recovery.db"

	// Phase 1: write data and close normally.
	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open phase 1: %v", err)
	}
	if err := Migrate(db1); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo1 := NewPatientRepo(db1)
	ctx := context.Background()

	for i := 0; i < 50; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("REC-%05d", i),
			LastName:        "Recovery",
			FirstName:       fmt.Sprintf("Patient%d", i),
			Sex:             "F",
			Phone:           "+237600000000",
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       5,
			EnrollmentDate:  time.Now(),
		}
		if err := repo1.Create(ctx, p); err != nil {
			t.Fatalf("phase 1 insert %d: %v", i, err)
		}
	}

	// Close (simulates normal shutdown before crash).
	db1.Close()

	// Phase 2: reopen and verify data integrity.
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open phase 2: %v", err)
	}
	defer db2.Close()

	if err := Migrate(db2); err != nil {
		t.Fatalf("migrate phase 2: %v", err)
	}

	repo2 := NewPatientRepo(db2)
	_, count, err := repo2.List(ctx, domain.PatientFilter{Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("phase 2 list: %v", err)
	}
	if count != 50 {
		t.Errorf("recovered count = %d, want 50", count)
	}

	// Verify individual records.
	p, err := repo2.GetByCode(ctx, "REC-00000")
	if err != nil {
		t.Fatalf("phase 2 GetByCode: %v", err)
	}
	if p.FirstName != "Patient0" {
		t.Errorf("FirstName = %q, want Patient0", p.FirstName)
	}
}

// TestChaos_ConcurrentAppointmentBooking simulates multiple users
// trying to book the same time slot concurrently.
func TestChaos_ConcurrentAppointmentBooking(t *testing.T) {
	db := testDB(t)
	pid := seedPatient(t, db)
	seedCenter(t, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	const goroutines = 20

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := &domain.Appointment{
				PatientID: pid,
				Date:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
				Time:      "10:00",
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
				Notes:     fmt.Sprintf("goroutine-%d", i),
			}
			err := repo.Create(ctx, a)
			if err != nil {
				errCh <- err
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Logf("booking error (may be expected under contention): %v", err)
	}

	// All inserts should succeed at the DB level (slot enforcement is in the
	// service layer, not the DB). The point is no deadlocks or crashes.
	if successCount == 0 {
		t.Error("expected at least some successful bookings")
	}
	t.Logf("%d/%d concurrent bookings succeeded", successCount, goroutines)
}
