package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

func seedUser(t *testing.T, db *DB) int64 {
	t.Helper()
	repo := NewUserRepo(db)
	u := &domain.User{
		Username:     "test",
		PasswordHash: "hash",
		FullName:     "Test",
		Role:         domain.RoleAdmin,
		Status:       domain.UserActive,
	}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u.ID
}

func TestSessionRepo_CreateAndGet(t *testing.T) {
	db := testDB(t)
	uid := seedUser(t, db)
	repo := NewSessionRepo(db)
	ctx := context.Background()

	s := &domain.Session{
		Token:     "abc123",
		UserID:    uid,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	}
	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByToken(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetByToken: %v", err)
	}
	if got.UserID != uid {
		t.Errorf("UserID = %d, want %d", got.UserID, uid)
	}
}

func TestSessionRepo_GetByToken_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewSessionRepo(db)

	_, err := repo.GetByToken(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("got %v, want ErrSessionExpired", err)
	}
}

func TestSessionRepo_DeleteByToken(t *testing.T) {
	db := testDB(t)
	uid := seedUser(t, db)
	repo := NewSessionRepo(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.Session{
		Token:     "todelete",
		UserID:    uid,
		ExpiresAt: time.Now().Add(time.Hour),
	})

	if err := repo.DeleteByToken(ctx, "todelete"); err != nil {
		t.Fatalf("DeleteByToken: %v", err)
	}

	_, err := repo.GetByToken(ctx, "todelete")
	if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("session should be gone, got %v", err)
	}
}

func TestSessionRepo_DeleteExpired(t *testing.T) {
	db := testDB(t)
	uid := seedUser(t, db)
	repo := NewSessionRepo(db)
	ctx := context.Background()

	// One expired, one valid.
	repo.Create(ctx, &domain.Session{
		Token:     "expired",
		UserID:    uid,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	})
	repo.Create(ctx, &domain.Session{
		Token:     "valid",
		UserID:    uid,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	if err := repo.DeleteExpired(ctx); err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}

	_, err := repo.GetByToken(ctx, "expired")
	if !errors.Is(err, domain.ErrSessionExpired) {
		t.Error("expired session should be gone")
	}

	_, err = repo.GetByToken(ctx, "valid")
	if err != nil {
		t.Errorf("valid session should still exist: %v", err)
	}
}
