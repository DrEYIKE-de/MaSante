package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/masante/masante/domain"
)

// ReminderRepo implements domain.ReminderRepository backed by SQLite.
type ReminderRepo struct {
	db *sql.DB
}

// NewReminderRepo returns a new ReminderRepo.
func NewReminderRepo(db *DB) *ReminderRepo {
	return &ReminderRepo{db: db.conn}
}

// Create inserts a new reminder.
func (r *ReminderRepo) Create(ctx context.Context, rem *domain.Reminder) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO reminders (appointment_id, patient_id, channel, type, message, status, scheduled_at)
		 VALUES (?,?,?,?,?,?,?)`,
		rem.AppointmentID, rem.PatientID, rem.Channel, rem.Type,
		rem.Message, rem.Status, rem.ScheduledAt.UTC().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return fmt.Errorf("insert reminder: %w", err)
	}
	id, _ := res.LastInsertId()
	rem.ID = id
	return nil
}

// GetByID returns a reminder by primary key.
func (r *ReminderRepo) GetByID(ctx context.Context, id int64) (*domain.Reminder, error) {
	rem := &domain.Reminder{}
	var scheduled, sentAt, created *string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, appointment_id, patient_id, channel, type, message, status,
		        scheduled_at, sent_at, provider_id, error_message, retry_count, created_at
		 FROM reminders WHERE id = ?`, id,
	).Scan(&rem.ID, &rem.AppointmentID, &rem.PatientID, &rem.Channel, &rem.Type,
		&rem.Message, &rem.Status, &scheduled, &sentAt, &rem.ProviderID,
		&rem.ErrorMessage, &rem.RetryCount, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrReminderNotFound
	}
	if err != nil {
		return nil, err
	}
	if scheduled != nil {
		rem.ScheduledAt = parseTimeVal(*scheduled)
	}
	rem.SentAt = parseTime(sentAt)
	if created != nil {
		rem.CreatedAt = parseTimeVal(*created)
	}
	return rem, nil
}

// Update persists changes to a reminder (status, sent_at, error, retry).
func (r *ReminderRepo) Update(ctx context.Context, rem *domain.Reminder) error {
	var sentStr *string
	if rem.SentAt != nil {
		s := rem.SentAt.UTC().Format("2006-01-02 15:04:05")
		sentStr = &s
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE reminders SET status=?, sent_at=?, provider_id=?, error_message=?, retry_count=?
		 WHERE id=?`,
		rem.Status, sentStr, rem.ProviderID, rem.ErrorMessage, rem.RetryCount, rem.ID,
	)
	return err
}

// ListPending returns reminders with status 'planifie' or 'echec' (for retry).
func (r *ReminderRepo) ListPending(ctx context.Context) ([]domain.Reminder, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.appointment_id, r.patient_id, r.channel, r.type, r.message, r.status,
		        r.scheduled_at, r.sent_at, r.provider_id, r.error_message, r.retry_count, r.created_at,
		        COALESCE(p.first_name || ' ' || p.last_name, '')
		 FROM reminders r
		 LEFT JOIN patients p ON p.id = r.patient_id
		 WHERE r.status IN ('planifie','echec')
		 ORDER BY r.scheduled_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ListByAppointment returns all reminders for a given appointment.
func (r *ReminderRepo) ListByAppointment(ctx context.Context, appointmentID int64) ([]domain.Reminder, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.appointment_id, r.patient_id, r.channel, r.type, r.message, r.status,
		        r.scheduled_at, r.sent_at, r.provider_id, r.error_message, r.retry_count, r.created_at,
		        COALESCE(p.first_name || ' ' || p.last_name, '')
		 FROM reminders r
		 LEFT JOIN patients p ON p.id = r.patient_id
		 WHERE r.appointment_id = ?
		 ORDER BY r.created_at`, appointmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ListByPatient returns all reminders for a given patient.
func (r *ReminderRepo) ListByPatient(ctx context.Context, patientID int64) ([]domain.Reminder, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.appointment_id, r.patient_id, r.channel, r.type, r.message, r.status,
		        r.scheduled_at, r.sent_at, r.provider_id, r.error_message, r.retry_count, r.created_at,
		        COALESCE(p.first_name || ' ' || p.last_name, '')
		 FROM reminders r
		 LEFT JOIN patients p ON p.id = r.patient_id
		 WHERE r.patient_id = ?
		 ORDER BY r.created_at DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// Stats returns delivery metrics across all reminders.
func (r *ReminderRepo) Stats(ctx context.Context) (domain.ReminderStats, error) {
	var stats domain.ReminderStats
	var total, sent, delivered, failed, pending int

	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM reminders GROUP BY status`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return stats, err
		}
		total += count
		switch domain.ReminderStatus(status) {
		case domain.ReminderSent:
			sent += count
		case domain.ReminderDelivered:
			delivered += count
		case domain.ReminderFailed:
			failed += count
		case domain.ReminderScheduled:
			pending += count
		}
	}

	if total > 0 {
		stats.DeliveryRate = float64(sent+delivered) / float64(total) * 100
	}
	stats.ConfirmRate = 0 // would require patient response tracking
	stats.PendingCount = pending
	stats.FailedCount = failed
	return stats, nil
}

// ListTemplates returns all message templates.
func (r *ReminderRepo) ListTemplates(ctx context.Context) ([]domain.MessageTemplate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, channel, body, language, is_active, created_at, updated_at
		 FROM message_templates ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.MessageTemplate
	for rows.Next() {
		var t domain.MessageTemplate
		var created, updated string
		if err := rows.Scan(&t.ID, &t.Name, &t.Channel, &t.Body, &t.Language,
			&t.IsActive, &created, &updated); err != nil {
			return nil, err
		}
		t.CreatedAt = parseTimeVal(created)
		t.UpdatedAt = parseTimeVal(updated)
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// UpdateTemplate modifies a message template.
func (r *ReminderRepo) UpdateTemplate(ctx context.Context, t *domain.MessageTemplate) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE message_templates SET body=?, is_active=?, updated_at=datetime('now') WHERE id=?`,
		t.Body, t.IsActive, t.ID)
	return err
}

func (r *ReminderRepo) scanRows(rows *sql.Rows) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	for rows.Next() {
		var rem domain.Reminder
		var scheduled, sentAt, created *string
		if err := rows.Scan(&rem.ID, &rem.AppointmentID, &rem.PatientID, &rem.Channel,
			&rem.Type, &rem.Message, &rem.Status, &scheduled, &sentAt,
			&rem.ProviderID, &rem.ErrorMessage, &rem.RetryCount, &created,
			&rem.PatientName); err != nil {
			return nil, err
		}
		if scheduled != nil {
			rem.ScheduledAt = parseTimeVal(*scheduled)
		}
		rem.SentAt = parseTime(sentAt)
		if created != nil {
			rem.CreatedAt = parseTimeVal(*created)
		}
		reminders = append(reminders, rem)
	}
	return reminders, rows.Err()
}
