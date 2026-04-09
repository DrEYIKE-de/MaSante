package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/masante/masante/domain"
)

func parseTime(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", "2006-01-02"} {
		if t, err := time.Parse(layout, *s); err == nil {
			return &t
		}
	}
	return nil
}

func parseTimeVal(s string) time.Time {
	t := parseTime(&s)
	if t == nil {
		return time.Time{}
	}
	return *t
}

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db.conn}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, full_name, email, phone, role, title, status, must_change_pwd)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.FullName, u.Email, u.Phone, u.Role, u.Title, u.Status, u.MustChangePwd,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u := &domain.User{}
	var lastLogin, created, updated *string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, full_name, email, phone, role, title, status, must_change_pwd, last_login_at, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Email, &u.Phone, &u.Role, &u.Title, &u.Status, &u.MustChangePwd, &lastLogin, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.LastLoginAt = parseTime(lastLogin)
	if created != nil {
		u.CreatedAt = parseTimeVal(*created)
	}
	if updated != nil {
		u.UpdatedAt = parseTimeVal(*updated)
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	u := &domain.User{}
	var lastLogin, created, updated *string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, full_name, email, phone, role, title, status, must_change_pwd, last_login_at, created_at, updated_at
		 FROM users WHERE username = ? COLLATE NOCASE`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Email, &u.Phone, &u.Role, &u.Title, &u.Status, &u.MustChangePwd, &lastLogin, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.LastLoginAt = parseTime(lastLogin)
	if created != nil {
		u.CreatedAt = parseTimeVal(*created)
	}
	if updated != nil {
		u.UpdatedAt = parseTimeVal(*updated)
	}
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET full_name=?, email=?, phone=?, role=?, title=?, status=?, must_change_pwd=?, password_hash=?, updated_at=datetime('now')
		 WHERE id=?`,
		u.FullName, u.Email, u.Phone, u.Role, u.Title, u.Status, u.MustChangePwd, u.PasswordHash, u.ID,
	)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET status='desactive', updated_at=datetime('now') WHERE id=?`, id)
	return err
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, full_name, email, phone, role, title, status, last_login_at, created_at
		 FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var lastLogin, created *string
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.Phone, &u.Role, &u.Title, &u.Status, &lastLogin, &created); err != nil {
			return nil, err
		}
		u.LastLoginAt = parseTime(lastLogin)
		if created != nil {
			u.CreatedAt = parseTimeVal(*created)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339), id)
	return err
}
