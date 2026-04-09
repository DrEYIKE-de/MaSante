package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/masante/masante/domain"
)

func TestUserRepo_CreateAndGetByID(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	u := &domain.User{
		Username:     "admin",
		PasswordHash: "hash",
		FullName:     "Dr. Mbarga",
		Email:        "mbarga@test.cm",
		Role:         domain.RoleAdmin,
		Status:       domain.UserActive,
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("ID not set after Create")
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Username != "admin" {
		t.Errorf("Username = %q, want admin", got.Username)
	}
	if got.FullName != "Dr. Mbarga" {
		t.Errorf("FullName = %q, want Dr. Mbarga", got.FullName)
	}
	if got.Role != domain.RoleAdmin {
		t.Errorf("Role = %q, want admin", got.Role)
	}
}

func TestUserRepo_GetByUsername_CaseInsensitive(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.User{
		Username:     "Admin",
		PasswordHash: "hash",
		FullName:     "Admin",
		Role:         domain.RoleAdmin,
		Status:       domain.UserActive,
	})

	got, err := repo.GetByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if got.Username != "Admin" {
		t.Errorf("Username = %q, want Admin", got.Username)
	}
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepo(db)

	_, err := repo.GetByID(context.Background(), 999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("got %v, want ErrUserNotFound", err)
	}
}

func TestUserRepo_Update(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	u := &domain.User{
		Username:     "test",
		PasswordHash: "hash",
		FullName:     "Original",
		Role:         domain.RoleMedecin,
		Status:       domain.UserActive,
	}
	repo.Create(ctx, u)

	u.FullName = "Updated"
	u.Role = domain.RoleInfirmier
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, u.ID)
	if got.FullName != "Updated" {
		t.Errorf("FullName = %q, want Updated", got.FullName)
	}
	if got.Role != domain.RoleInfirmier {
		t.Errorf("Role = %q, want infirmier", got.Role)
	}
}

func TestUserRepo_Delete_SoftDisable(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	u := &domain.User{
		Username:     "toDelete",
		PasswordHash: "hash",
		FullName:     "To Delete",
		Role:         domain.RoleASC,
		Status:       domain.UserActive,
	}
	repo.Create(ctx, u)

	if err := repo.Delete(ctx, u.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, _ := repo.GetByID(ctx, u.ID)
	if got.Status != domain.UserDisabled {
		t.Errorf("Status = %q, want desactive", got.Status)
	}
}

func TestUserRepo_List(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	for _, name := range []string{"alice", "bob", "charlie"} {
		repo.Create(ctx, &domain.User{
			Username:     name,
			PasswordHash: "hash",
			FullName:     name,
			Role:         domain.RoleMedecin,
			Status:       domain.UserActive,
		})
	}

	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("got %d users, want 3", len(users))
	}
}
