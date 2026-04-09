package sqlite

import (
	"context"
	"database/sql"

	"github.com/masante/masante/domain"
)

type AuditRepo struct {
	db *sql.DB
}

func NewAuditRepo(db *DB) *AuditRepo {
	return &AuditRepo{db: db.conn}
}

func (r *AuditRepo) Log(ctx context.Context, e *domain.AuditEntry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_log (user_id, action, entity_type, entity_id, details, ip_address) VALUES (?, ?, ?, ?, ?, ?)`,
		e.UserID, e.Action, e.EntityType, e.EntityID, e.Details, e.IPAddress,
	)
	return err
}

func (r *AuditRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]domain.AuditEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, action, entity_type, entity_id, details, ip_address, created_at
		 FROM audit_log WHERE user_id = ? ORDER BY created_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Action, &e.EntityType, &e.EntityID, &e.Details, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
