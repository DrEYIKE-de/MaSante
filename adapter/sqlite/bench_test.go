package sqlite

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

// benchDB creates a database for benchmarking, shared across sub-benchmarks.
func benchDB(b *testing.B) *DB {
	b.Helper()
	dir := b.TempDir()
	db, err := Open(dir + "/bench.db")
	if err != nil {
		b.Fatalf("open: %v", err)
	}
	if err := Migrate(db); err != nil {
		b.Fatalf("migrate: %v", err)
	}
	b.Cleanup(func() { db.Close() })
	return db
}

func benchSeedUser(b *testing.B, db *DB) int64 {
	b.Helper()
	repo := NewUserRepo(db)
	u := &domain.User{
		Username:     "benchuser",
		PasswordHash: "hash",
		FullName:     "Bench User",
		Role:         domain.RoleAdmin,
		Status:       domain.UserActive,
	}
	if err := repo.Create(context.Background(), u); err != nil {
		b.Fatalf("seed user: %v", err)
	}
	return u.ID
}

func benchSeedPatient(b *testing.B, db *DB) int64 {
	b.Helper()
	repo := NewPatientRepo(db)
	p := &domain.Patient{
		Code:            "BN-2026-00001",
		LastName:        "Bench",
		FirstName:       "Patient",
		Sex:             "M",
		Phone:           "+237600000000",
		Language:        "fr",
		ReminderChannel: domain.ChannelSMS,
		Status:          domain.PatientActive,
		RiskScore:       5,
		EnrollmentDate:  time.Now(),
	}
	if err := repo.Create(context.Background(), p); err != nil {
		b.Fatalf("seed patient: %v", err)
	}
	return p.ID
}

func benchSeedCenter(b *testing.B, db *DB) {
	b.Helper()
	repo := NewCenterRepo(db)
	repo.Create(context.Background(), &domain.Center{
		Name: "Bench Center", Type: domain.CenterClinic, Country: "Cameroun", City: "Douala",
	})
}

// BenchmarkPatientCreate measures the throughput of patient insertions.
func BenchmarkPatientCreate(b *testing.B) {
	db := benchDB(b)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("BC-%07d", i),
			LastName:        "Benchmark",
			FirstName:       fmt.Sprintf("Patient%d", i),
			Sex:             "F",
			Phone:           "+237600000000",
			District:        "Akwa",
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       5,
			EnrollmentDate:  time.Now(),
		}
		if err := repo.Create(ctx, p); err != nil {
			b.Fatalf("create: %v", err)
		}
	}
}

// BenchmarkPatientSearch measures search performance across a populated database.
func BenchmarkPatientSearch(b *testing.B) {
	db := benchDB(b)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	// Pre-populate with 500 patients.
	for i := 0; i < 500; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("BS-%05d", i),
			LastName:        fmt.Sprintf("SearchLast_%d", i),
			FirstName:       fmt.Sprintf("SearchFirst_%d", i),
			Sex:             []string{"M", "F"}[i%2],
			Phone:           fmt.Sprintf("+237%09d", i),
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       i % 11,
			EnrollmentDate:  time.Now(),
		}
		repo.Create(ctx, p)
	}

	queries := []string{"SearchLast_25", "BS-00100", "+237000000050", "SearchFirst_3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		if _, err := repo.Search(ctx, q, 10); err != nil {
			b.Fatalf("search: %v", err)
		}
	}
}

// BenchmarkPatientList measures listing performance with pagination.
func BenchmarkPatientList(b *testing.B) {
	db := benchDB(b)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	// Pre-populate with 500 patients.
	for i := 0; i < 500; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("BL-%05d", i),
			LastName:        fmt.Sprintf("ListLast_%d", i),
			FirstName:       fmt.Sprintf("ListFirst_%d", i),
			Sex:             "M",
			Phone:           "+237600000000",
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       5,
			EnrollmentDate:  time.Now(),
		}
		repo.Create(ctx, p)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page := (i % 25) + 1
		if _, _, err := repo.List(ctx, domain.PatientFilter{Page: page, PerPage: 20}); err != nil {
			b.Fatalf("list: %v", err)
		}
	}
}

// BenchmarkAppointmentCreate measures appointment insertion throughput.
func BenchmarkAppointmentCreate(b *testing.B) {
	db := benchDB(b)
	pid := benchSeedPatient(b, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i%365)
		a := &domain.Appointment{
			PatientID: pid,
			Date:      date,
			Time:      fmt.Sprintf("%02d:%02d", 8+(i%8), (i%2)*30),
			Type:      domain.TypeConsultation,
			Status:    domain.StatusConfirmed,
			Notes:     "benchmark appointment",
		}
		if err := repo.Create(ctx, a); err != nil {
			b.Fatalf("create: %v", err)
		}
	}
}

// BenchmarkSessionCreateDelete measures the full create+delete cycle for sessions.
func BenchmarkSessionCreateDelete(b *testing.B) {
	db := benchDB(b)
	uid := benchSeedUser(b, db)
	repo := NewSessionRepo(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := fmt.Sprintf("bench-sd-token-%d", i)
		s := &domain.Session{
			Token:     token,
			UserID:    uid,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IPAddress: "127.0.0.1",
			UserAgent: "benchmark",
		}
		if err := repo.Create(ctx, s); err != nil {
			b.Fatalf("create session: %v", err)
		}
		if err := repo.DeleteByToken(ctx, token); err != nil {
			b.Fatalf("delete session: %v", err)
		}
	}
}

// BenchmarkConcurrentReads measures read throughput under concurrent load
// using b.RunParallel.
func BenchmarkConcurrentReads(b *testing.B) {
	db := benchDB(b)
	repo := NewPatientRepo(db)
	ctx := context.Background()

	// Pre-populate with 200 patients.
	for i := 0; i < 200; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("CR-%05d", i),
			LastName:        fmt.Sprintf("ConcRead_%d", i),
			FirstName:       fmt.Sprintf("Patient_%d", i),
			Sex:             "M",
			Phone:           "+237600000000",
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       5,
			EnrollmentDate:  time.Now(),
		}
		repo.Create(ctx, p)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			switch i % 3 {
			case 0:
				repo.List(ctx, domain.PatientFilter{Page: (i % 10) + 1, PerPage: 20})
			case 1:
				repo.Search(ctx, fmt.Sprintf("ConcRead_%d", i%200), 10)
			case 2:
				repo.CountByStatus(ctx)
			}
		}
	})
}

// BenchmarkAppointmentListByDate measures date-based appointment lookups.
func BenchmarkAppointmentListByDate(b *testing.B) {
	db := benchDB(b)
	pid := benchSeedPatient(b, db)
	repo := NewAppointmentRepo(db)
	ctx := context.Background()

	// Pre-populate: 10 appointments per day for 30 days.
	for day := 0; day < 30; day++ {
		date := time.Date(2026, 4, 1+day, 0, 0, 0, 0, time.UTC)
		for slot := 0; slot < 10; slot++ {
			a := &domain.Appointment{
				PatientID: pid,
				Date:      date,
				Time:      fmt.Sprintf("%02d:00", 8+slot),
				Type:      domain.TypeConsultation,
				Status:    domain.StatusConfirmed,
			}
			repo.Create(ctx, a)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		date := time.Date(2026, 4, 1+(i%30), 0, 0, 0, 0, time.UTC)
		if _, err := repo.ListByDate(ctx, date); err != nil {
			b.Fatalf("list by date: %v", err)
		}
	}
}

// BenchmarkMixedWorkload simulates a realistic mix of reads and writes.
func BenchmarkMixedWorkload(b *testing.B) {
	db := benchDB(b)
	patientRepo := NewPatientRepo(db)
	ctx := context.Background()

	// Seed some data.
	for i := 0; i < 100; i++ {
		p := &domain.Patient{
			Code:            fmt.Sprintf("MW-%05d", i),
			LastName:        fmt.Sprintf("Mixed_%d", i),
			FirstName:       fmt.Sprintf("Work_%d", i),
			Sex:             "F",
			Phone:           "+237600000000",
			Language:        "fr",
			ReminderChannel: domain.ChannelSMS,
			Status:          domain.PatientActive,
			RiskScore:       5,
			EnrollmentDate:  time.Now(),
		}
		patientRepo.Create(ctx, p)
	}

	var mu sync.Mutex
	writeCounter := 100

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		localI := 0
		for pb.Next() {
			localI++
			switch localI % 5 {
			case 0: // 20% writes
				mu.Lock()
				writeCounter++
				code := fmt.Sprintf("MW-%05d", writeCounter)
				mu.Unlock()
				p := &domain.Patient{
					Code:            code,
					LastName:        "NewPatient",
					FirstName:       fmt.Sprintf("P_%d", localI),
					Sex:             "M",
					Phone:           "+237600000000",
					Language:        "fr",
					ReminderChannel: domain.ChannelSMS,
					Status:          domain.PatientActive,
					RiskScore:       5,
					EnrollmentDate:  time.Now(),
				}
				patientRepo.Create(ctx, p)
			default: // 80% reads
				patientRepo.List(ctx, domain.PatientFilter{Page: 1, PerPage: 20})
			}
		}
	})
}
