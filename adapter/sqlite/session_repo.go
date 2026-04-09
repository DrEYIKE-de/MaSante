package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/masante/masante/domain"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db.conn}
}

func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, expires_at, ip_address, user_agent) VALUES (?, ?, ?, ?, ?)`,
		s.Token, s.UserID, s.ExpiresAt.UTC().Format("2006-01-02 15:04:05"), s.IPAddress, s.UserAgent,
	)
	return err
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	s := &domain.Session{}
	var expires, created string
	err := r.db.QueryRowContext(ctx,
		`SELECT token, user_id, expires_at, ip_address, user_agent, created_at FROM sessions WHERE token = ?`, token,
	).Scan(&s.Token, &s.UserID, &expires, &s.IPAddress, &s.UserAgent, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrSessionExpired
	}
	if err != nil {
		return nil, err
	}
	s.ExpiresAt = parseTimeVal(expires)
	s.CreatedAt = parseTimeVal(created)
	return s, nil
}

func (r *SessionRepo) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (r *SessionRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < datetime('now')`)
	return err
}
