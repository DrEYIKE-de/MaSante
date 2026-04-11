package app

import (
	"context"
	"errors"
	"testing"

	"github.com/masante/masante/domain"
)

func newTestUserService() *UserService {
	users := &mockUserRepo{users: map[string]*domain.User{}}
	sessions := &mockSessionRepo{sessions: make(map[string]*domain.Session)}
	return NewUserService(users, sessions, mockHasher{}, &mockAudit{})
}

func TestUserService_Create(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	u := &domain.User{
		Username: "ngassa.m",
		FullName: "Ngassa Marie",
		Email:    "ngassa@mail.cm",
		Role:     domain.RoleASC,
		Status:   domain.UserActive,
	}

	if err := svc.Create(ctx, u, "temppass123", 1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == 0 {
		t.Error("ID should be set")
	}
	if !u.MustChangePwd {
		t.Error("MustChangePwd should be true for new users")
	}
	if u.PasswordHash == "" {
		t.Error("PasswordHash should be set")
	}
}

func TestUserService_Disable(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	// Create an admin (ID=1) and a medecin (ID=2).
	admin := &domain.User{Username: "admin", FullName: "Admin", Role: domain.RoleAdmin, Status: domain.UserActive}
	svc.Create(ctx, admin, "password123", 0)
	target := &domain.User{Username: "toremove", FullName: "To Remove", Role: domain.RoleMedecin, Status: domain.UserActive}
	svc.Create(ctx, target, "password123", admin.ID)

	// Admin disables the medecin.
	if err := svc.Disable(ctx, target.ID, admin.ID); err != nil {
		t.Fatalf("Disable: %v", err)
	}

	got, err := svc.GetByID(ctx, target.ID)
	if err != nil {
		t.Fatalf("GetByID after disable: %v", err)
	}
	if got.Status != domain.UserDisabled {
		t.Errorf("Status = %q, want desactive", got.Status)
	}
}

func TestUserService_Disable_CannotDisableSelf(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	admin := &domain.User{Username: "admin", FullName: "Admin", Role: domain.RoleAdmin, Status: domain.UserActive}
	svc.Create(ctx, admin, "password123", 0)

	err := svc.Disable(ctx, admin.ID, admin.ID)
	if err == nil {
		t.Fatal("expected error when disabling self")
	}
}

func TestUserService_Disable_CannotDisableLastAdmin(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	admin := &domain.User{Username: "admin", FullName: "Admin", Role: domain.RoleAdmin, Status: domain.UserActive}
	svc.Create(ctx, admin, "password123", 0)
	other := &domain.User{Username: "other", FullName: "Other", Role: domain.RoleMedecin, Status: domain.UserActive}
	svc.Create(ctx, other, "password123", admin.ID)

	// Other tries to disable the only admin.
	err := svc.Disable(ctx, admin.ID, other.ID)
	if err == nil {
		t.Fatal("expected error when disabling last admin")
	}
}

func TestUserService_ResetPassword(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	u := &domain.User{Username: "resetme", FullName: "Reset Me", Role: domain.RoleInfirmier, Status: domain.UserActive}
	svc.Create(ctx, u, "oldpass123", 1)

	if err := svc.ResetPassword(ctx, u.ID, "newpass456", 1); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	got, _ := svc.GetByID(ctx, u.ID)
	if !got.MustChangePwd {
		t.Error("MustChangePwd should be true after reset")
	}
	if got.PasswordHash != "hashed:newpass456" {
		t.Errorf("PasswordHash = %q, want hashed:newpass456", got.PasswordHash)
	}
}

func TestUserService_ChangePassword_OK(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	u := &domain.User{Username: "chgpwd", FullName: "Change Pwd", Role: domain.RoleMedecin, Status: domain.UserActive}
	svc.Create(ctx, u, "current123", 1)

	if err := svc.ChangePassword(ctx, u.ID, "current123", "newpass789"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}

	got, _ := svc.GetByID(ctx, u.ID)
	if got.MustChangePwd {
		t.Error("MustChangePwd should be false after self-change")
	}
}

func TestUserService_ChangePassword_WrongCurrent(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	u := &domain.User{Username: "wrongpwd", FullName: "Wrong Pwd", Role: domain.RoleMedecin, Status: domain.UserActive}
	svc.Create(ctx, u, "correct123", 1)

	err := svc.ChangePassword(ctx, u.ID, "wrong", "newpass789")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("got %v, want ErrInvalidPassword", err)
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	svc := newTestUserService()
	ctx := context.Background()

	u := &domain.User{Username: "profile", FullName: "Original", Role: domain.RoleMedecin, Status: domain.UserActive}
	svc.Create(ctx, u, "password123", 1)

	if err := svc.UpdateProfile(ctx, u.ID, "Updated Name", "new@mail.cm", "+237600000000"); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	got, _ := svc.GetByID(ctx, u.ID)
	if got.FullName != "Updated Name" {
		t.Errorf("FullName = %q, want Updated Name", got.FullName)
	}
	if got.Email != "new@mail.cm" {
		t.Errorf("Email = %q, want new@mail.cm", got.Email)
	}
}

func TestUserService_ResetPassword_NotFound(t *testing.T) {
	svc := newTestUserService()

	err := svc.ResetPassword(context.Background(), 999, "newpass", 1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("got %v, want ErrUserNotFound", err)
	}
}
