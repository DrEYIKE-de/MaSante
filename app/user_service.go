package app

import (
	"context"
	"fmt"

	"github.com/masante/masante/domain"
)

// UserService handles user CRUD and profile operations.
type UserService struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	hasher   domain.PasswordHasher
	audit    domain.AuditRepository
}

// NewUserService returns a new UserService.
func NewUserService(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	hasher domain.PasswordHasher,
	audit domain.AuditRepository,
) *UserService {
	return &UserService{users: users, sessions: sessions, hasher: hasher, audit: audit}
}

// Create adds a new user account. Only admins should call this.
func (s *UserService) Create(ctx context.Context, u *domain.User, password string, createdBy int64) error {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u.PasswordHash = hash
	u.MustChangePwd = true

	if err := s.users.Create(ctx, u); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &createdBy,
		Action:     "user.create",
		EntityType: "user",
		EntityID:   &u.ID,
		Details:    string(u.Role),
	})
	return nil
}

// GetByID returns a user by ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.users.List(ctx)
}

// Update modifies a user's profile fields and role.
func (s *UserService) Update(ctx context.Context, u *domain.User, updatedBy int64) error {
	if err := s.users.Update(ctx, u); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &updatedBy,
		Action:     "user.update",
		EntityType: "user",
		EntityID:   &u.ID,
	})
	return nil
}

// Disable soft-deletes a user and revokes all sessions.
func (s *UserService) Disable(ctx context.Context, id int64, disabledBy int64) error {
	if err := s.users.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.sessions.DeleteByUserID(ctx, id)
	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &disabledBy,
		Action:     "user.disable",
		EntityType: "user",
		EntityID:   &id,
	})
	return nil
}

// ResetPassword sets a new temporary password and forces change at next login.
func (s *UserService) ResetPassword(ctx context.Context, id int64, newPassword string, resetBy int64) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}

	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	u.PasswordHash = hash
	u.MustChangePwd = true
	if err := s.users.Update(ctx, u); err != nil {
		return err
	}

	// Revoke existing sessions so the user must re-login.
	_ = s.sessions.DeleteByUserID(ctx, id)

	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &resetBy,
		Action:     "user.reset_password",
		EntityType: "user",
		EntityID:   &id,
	})
	return nil
}

// ChangePassword lets a user change their own password.
func (s *UserService) ChangePassword(ctx context.Context, id int64, currentPwd, newPwd string) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.hasher.Verify(u.PasswordHash, currentPwd); err != nil {
		return domain.ErrInvalidPassword
	}

	hash, err := s.hasher.Hash(newPwd)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	u.PasswordHash = hash
	u.MustChangePwd = false
	if err := s.users.Update(ctx, u); err != nil {
		return err
	}
	// Revoke all existing sessions — the user must re-login.
	_ = s.sessions.DeleteByUserID(ctx, id)
	return nil
}

// UpdateProfile lets a user update their own non-sensitive fields.
func (s *UserService) UpdateProfile(ctx context.Context, id int64, fullName, email, phone string) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}

	u.FullName = fullName
	u.Email = email
	u.Phone = phone
	return s.users.Update(ctx, u)
}

// Activity returns recent audit entries for a user.
func (s *UserService) Activity(ctx context.Context, userID int64, limit int) ([]domain.AuditEntry, error) {
	return s.audit.ListByUser(ctx, userID, limit)
}
