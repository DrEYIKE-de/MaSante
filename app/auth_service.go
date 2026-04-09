package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/masante/masante/domain"
)

const sessionDuration = 24 * time.Hour

type AuthService struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	hasher   domain.PasswordHasher
	audit    domain.AuditRepository
}

func NewAuthService(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	hasher domain.PasswordHasher,
	audit domain.AuditRepository,
) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		hasher:   hasher,
		audit:    audit,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password, ip, ua string) (*domain.Session, *domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, domain.ErrInvalidPassword
	}

	if user.Status != domain.UserActive {
		return nil, nil, domain.ErrUnauthorized
	}

	if err := s.hasher.Verify(user.PasswordHash, password); err != nil {
		return nil, nil, domain.ErrInvalidPassword
	}

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

	_ = s.users.UpdateLastLogin(ctx, user.ID)
	_ = s.audit.Log(ctx, &domain.AuditEntry{
		UserID:     &user.ID,
		Action:     "auth.login",
		EntityType: "user",
		EntityID:   &user.ID,
		IPAddress:  ip,
	})

	return session, user, nil
}

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

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessions.DeleteByToken(ctx, token)
}

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
