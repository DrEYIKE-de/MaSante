package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

// --- mocks ---

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(_ context.Context, u *domain.User) error {
	u.ID = int64(len(m.users) + 1)
	m.users[u.Username] = u
	return nil
}
func (m *mockUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}
func (m *mockUserRepo) Update(_ context.Context, _ *domain.User) error   { return nil }
func (m *mockUserRepo) Delete(_ context.Context, _ int64) error          { return nil }
func (m *mockUserRepo) List(_ context.Context) ([]domain.User, error)    { return nil, nil }
func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ int64) error { return nil }

type mockSessionRepo struct {
	sessions map[string]*domain.Session
}

func (m *mockSessionRepo) Create(_ context.Context, s *domain.Session) error {
	m.sessions[s.Token] = s
	return nil
}
func (m *mockSessionRepo) GetByToken(_ context.Context, token string) (*domain.Session, error) {
	s, ok := m.sessions[token]
	if !ok {
		return nil, domain.ErrSessionExpired
	}
	return s, nil
}
func (m *mockSessionRepo) DeleteByToken(_ context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}
func (m *mockSessionRepo) DeleteByUserID(_ context.Context, userID int64) error {
	for k, s := range m.sessions {
		if s.UserID == userID {
			delete(m.sessions, k)
		}
	}
	return nil
}
func (m *mockSessionRepo) DeleteExpired(_ context.Context) error { return nil }

type mockHasher struct{}

func (mockHasher) Hash(password string) (string, error)        { return "hashed:" + password, nil }
func (mockHasher) Verify(hash, password string) error {
	if hash == "hashed:"+password {
		return nil
	}
	return errors.New("mismatch")
}

type mockAudit struct{ entries []domain.AuditEntry }

func (m *mockAudit) Log(_ context.Context, e *domain.AuditEntry) error {
	m.entries = append(m.entries, *e)
	return nil
}
func (m *mockAudit) ListByUser(_ context.Context, _ int64, _ int) ([]domain.AuditEntry, error) {
	return m.entries, nil
}

// --- helpers ---

func newTestAuthService() (*AuthService, *mockSessionRepo) {
	users := &mockUserRepo{users: map[string]*domain.User{
		"admin": {
			ID:           1,
			Username:     "admin",
			PasswordHash: "hashed:secret",
			FullName:     "Admin",
			Role:         domain.RoleAdmin,
			Status:       domain.UserActive,
		},
		"disabled": {
			ID:           2,
			Username:     "disabled",
			PasswordHash: "hashed:secret",
			FullName:     "Disabled User",
			Role:         domain.RoleMedecin,
			Status:       domain.UserDisabled,
		},
	}}
	sessions := &mockSessionRepo{sessions: make(map[string]*domain.Session)}
	audit := &mockAudit{}
	svc := NewAuthService(users, sessions, mockHasher{}, audit)
	return svc, sessions
}

// --- tests ---

func TestAuthService_Login_OK(t *testing.T) {
	svc, sessions := newTestAuthService()

	sess, user, err := svc.Login(context.Background(), "admin", "secret", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("got username %q, want admin", user.Username)
	}
	if sess.Token == "" {
		t.Error("session token is empty")
	}
	if len(sessions.sessions) != 1 {
		t.Errorf("got %d sessions, want 1", len(sessions.sessions))
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, _ := newTestAuthService()

	_, _, err := svc.Login(context.Background(), "admin", "wrong", "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("got %v, want ErrInvalidPassword", err)
	}
}

func TestAuthService_Login_UnknownUser(t *testing.T) {
	svc, _ := newTestAuthService()

	_, _, err := svc.Login(context.Background(), "nobody", "secret", "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("got %v, want ErrInvalidPassword", err)
	}
}

func TestAuthService_Login_DisabledUser(t *testing.T) {
	svc, _ := newTestAuthService()

	_, _, err := svc.Login(context.Background(), "disabled", "secret", "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("got %v, want ErrUnauthorized", err)
	}
}

func TestAuthService_Authenticate_OK(t *testing.T) {
	svc, _ := newTestAuthService()

	sess, _, _ := svc.Login(context.Background(), "admin", "secret", "127.0.0.1", "test")

	user, err := svc.Authenticate(context.Background(), sess.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("got user ID %d, want 1", user.ID)
	}
}

func TestAuthService_Authenticate_ExpiredSession(t *testing.T) {
	svc, sessions := newTestAuthService()

	sessions.sessions["expired"] = &domain.Session{
		Token:     "expired",
		UserID:    1,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	_, err := svc.Authenticate(context.Background(), "expired")
	if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("got %v, want ErrSessionExpired", err)
	}
}

func TestAuthService_Authenticate_InvalidToken(t *testing.T) {
	svc, _ := newTestAuthService()

	_, err := svc.Authenticate(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("got %v, want ErrSessionExpired", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	svc, sessions := newTestAuthService()

	sess, _, _ := svc.Login(context.Background(), "admin", "secret", "127.0.0.1", "test")
	if len(sessions.sessions) != 1 {
		t.Fatal("expected 1 session after login")
	}

	if err := svc.Logout(context.Background(), sess.Token); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if len(sessions.sessions) != 0 {
		t.Error("session not deleted after logout")
	}
}
