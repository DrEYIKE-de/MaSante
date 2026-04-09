package sqlite

import (
	"context"
	"testing"

	"github.com/masante/masante/domain"
)

func TestCenterRepo_CreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewCenterRepo(db)
	ctx := context.Background()

	c := &domain.Center{
		Name:    "Hopital Laquintinie",
		Type:    domain.CenterHospital,
		Country: "Cameroun",
		City:    "Douala",
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Hopital Laquintinie" {
		t.Errorf("Name = %q, want Hopital Laquintinie", got.Name)
	}
	if got.SetupComplete {
		t.Error("SetupComplete should be false initially")
	}
}

func TestCenterRepo_IsSetupDone_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewCenterRepo(db)

	done, err := repo.IsSetupDone(context.Background())
	if err != nil {
		t.Fatalf("IsSetupDone: %v", err)
	}
	if done {
		t.Error("setup should not be done on empty database")
	}
}

func TestCenterRepo_CompleteSetup(t *testing.T) {
	db := testDB(t)
	repo := NewCenterRepo(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.Center{
		Name:    "Test",
		Type:    domain.CenterClinic,
		Country: "Cameroun",
		City:    "Douala",
	})

	if err := repo.CompleteSetup(ctx); err != nil {
		t.Fatalf("CompleteSetup: %v", err)
	}

	done, _ := repo.IsSetupDone(ctx)
	if !done {
		t.Error("setup should be done after CompleteSetup")
	}
}

func TestCenterRepo_Update(t *testing.T) {
	db := testDB(t)
	repo := NewCenterRepo(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.Center{
		Name:    "Original",
		Type:    domain.CenterClinic,
		Country: "Cameroun",
		City:    "Douala",
	})

	c, _ := repo.Get(ctx)
	c.StartTime = "07:30"
	c.SlotDuration = 45
	c.MaxPatientsDay = 50
	if err := repo.Update(ctx, c); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.Get(ctx)
	if got.StartTime != "07:30" {
		t.Errorf("StartTime = %q, want 07:30", got.StartTime)
	}
	if got.SlotDuration != 45 {
		t.Errorf("SlotDuration = %d, want 45", got.SlotDuration)
	}
}
