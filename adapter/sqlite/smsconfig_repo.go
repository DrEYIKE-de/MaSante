package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/masante/masante/domain"
)

type SMSConfigRepo struct {
	db *sql.DB
}

func NewSMSConfigRepo(db *DB) *SMSConfigRepo {
	return &SMSConfigRepo{db: db.conn}
}

func (r *SMSConfigRepo) Get(ctx context.Context) (*domain.SMSConfig, error) {
	c := &domain.SMSConfig{}
	err := r.db.QueryRowContext(ctx,
		`SELECT enabled, provider, api_key, api_secret, sender_id,
		        reminder_j7, reminder_j2, reminder_j0, reminder_late, late_delay_days, updated_at
		 FROM sms_config WHERE id = 1`,
	).Scan(&c.Enabled, &c.Provider, &c.APIKey, &c.APISecret, &c.SenderID,
		&c.ReminderJ7, &c.ReminderJ2, &c.ReminderJ0, &c.ReminderLate, &c.LateDelayDays, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return &domain.SMSConfig{}, nil
	}
	return c, err
}

func (r *SMSConfigRepo) Save(ctx context.Context, c *domain.SMSConfig) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sms_config
		 (id, enabled, provider, api_key, api_secret, sender_id, reminder_j7, reminder_j2, reminder_j0, reminder_late, late_delay_days)
		 VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Enabled, c.Provider, c.APIKey, c.APISecret, c.SenderID,
		c.ReminderJ7, c.ReminderJ2, c.ReminderJ0, c.ReminderLate, c.LateDelayDays,
	)
	return err
}
