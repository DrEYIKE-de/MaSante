package domain

import (
	"context"
	"errors"
	"time"
)

// Role defines the access level of a user.
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleMedecin   Role = "medecin"
	RoleInfirmier Role = "infirmier"
	RoleASC       Role = "asc"
)

// UserStatus represents account availability.
type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserOnLeave  UserStatus = "conge"
	UserDisabled UserStatus = "desactive"
)

// User is anyone who can log in to the system.
type User struct {
	ID            int64
	Username      string
	PasswordHash  string
	FullName      string
	Email         string
	Phone         string
	Role          Role
	Title         string
	Status        UserStatus
	MustChangePwd bool
	LastLoginAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CanAccess reports whether the user's role grants access to resource.
func (u *User) CanAccess(resource string) bool {
	perms, ok := rolePermissions[u.Role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == "*" || p == resource {
			return true
		}
	}
	return false
}

var rolePermissions = map[Role][]string{
	RoleAdmin:     {"*"},
	RoleMedecin:   {"dashboard", "patient", "appointment", "calendar", "reminder.read", "export", "profile"},
	RoleInfirmier: {"dashboard", "patient", "appointment", "calendar", "reminder", "profile"},
	RoleASC:       {"asc", "patient.read", "appointment.read", "profile"},
}

// Session represents an authenticated user session.
type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	CreatedAt time.Time
}

// UserRepository is a driven port for user persistence.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]User, error)
	UpdateLastLogin(ctx context.Context, id int64) error
}

// SessionRepository is a driven port for session persistence.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) error
}

// PasswordHasher is a driven port for password hashing.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

var (
	ErrUserNotFound    = errors.New("utilisateur introuvable")
	ErrUsernameTaken   = errors.New("identifiant deja utilise")
	ErrInvalidPassword = errors.New("mot de passe incorrect")
	ErrSessionExpired  = errors.New("session expiree")
	ErrUnauthorized    = errors.New("acces non autorise")
)
