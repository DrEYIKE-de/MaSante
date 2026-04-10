package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/masante/masante/domain"
)

const (
	sessionDuration    = 24 * time.Hour
	maxFailedAttempts  = 5
	lockoutDuration    = 15 * time.Minute
)

// AuthService handles login, logout, and session management.
type AuthService struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	hasher   domain.PasswordHasher
	audit    domain.AuditRepository
}

// NewAuthService returns a new AuthService.
func NewAuthService(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	hasher domain.PasswordHasher,
	audit domain.AuditRepository,
) *AuthService {
	return &AuthService{users: users, sessions: sessions, hasher: hasher, audit: audit}
}

// Login authenticates a user and creates a session.
// Returns ErrAccountLocked after 5 failed attempts (15 min lockout).
func (s *AuthService) Login(ctx context.Context, username, password, ip, ua string) (*domain.Session, *domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, domain.ErrInvalidPassword
	}

	if user.Status != domain.UserActive {
		return nil, nil, domain.ErrUnauthorized
	}

	// Check lockout.
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, nil, domain.ErrAccountLocked
	}

	// Clear expired lockout.
	if user.LockedUntil != nil && time.Now().After(*user.LockedUntil) {
		_ = s.users.ResetFailedAttempts(ctx, user.ID)
		user.FailedAttempts = 0
		user.LockedUntil = nil
	}

	if err := s.hasher.Verify(user.PasswordHash, password); err != nil {
		// Record failed attempt.
		_ = s.users.IncrementFailedAttempts(ctx, user.ID)
		if user.FailedAttempts+1 >= maxFailedAttempts {
			lockUntil := time.Now().Add(lockoutDuration)
			_ = s.users.LockAccount(ctx, user.ID, lockUntil)
			return nil, nil, domain.ErrAccountLocked
		}
		return nil, nil, domain.ErrInvalidPassword
	}

	// Success — reset failed attempts.
	token, err := generateToken()
	if err != nil {
		return nil, nil, err
	}

	session := &domain.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(sessionDuration),
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	_ = s.users.UpdateLastLogin(ctx, user.ID) // also resets failed_attempts
	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &user.ID,
		Action:     "auth.login",
		EntityType: "user",
		EntityID:   &user.ID,
		IPAddress:  ip,
	})

	return session, user, nil
}

// Authenticate validates a session token and returns the user.
func (s *AuthService) Authenticate(ctx context.Context, token string) (*domain.User, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessions.DeleteByToken(ctx, token)
		return nil, domain.ErrSessionExpired
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	if user.Status != domain.UserActive {
		_ = s.sessions.DeleteByToken(ctx, token)
		return nil, domain.ErrUnauthorized
	}

	return user, nil
}

// Logout removes a session.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessions.DeleteByToken(ctx, token)
}

// CleanExpiredSessions removes all expired sessions from the store.
func (s *AuthService) CleanExpiredSessions(ctx context.Context) error {
	return s.sessions.DeleteExpired(ctx)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
