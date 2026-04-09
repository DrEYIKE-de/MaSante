package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/masante/masante/domain"
)

// AppointmentRepo implements domain.AppointmentRepository backed by SQLite.
type AppointmentRepo struct {
	db *sql.DB
}

// NewAppointmentRepo returns a new AppointmentRepo.
func NewAppointmentRepo(db *DB) *AppointmentRepo {
	return &AppointmentRepo{db: db.conn}
}

// Create inserts a new appointment. It sets a.ID on success.
func (r *AppointmentRepo) Create(ctx context.Context, a *domain.Appointment) error {
	var freq *string
	if a.FollowUpFreq != nil {
		s := string(*a.FollowUpFreq)
		freq = &s
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO appointments (patient_id, user_id, date, time, type, status, notes, follow_up_freq, created_by)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		a.PatientID, a.UserID, a.Date.Format("2006-01-02"), a.Time,
		a.Type, a.Status, a.Notes, freq, a.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert appointment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

// GetByID returns an appointment by primary key.
func (r *AppointmentRepo) GetByID(ctx context.Context, id int64) (*domain.Appointment, error) {
	a, err := r.scanOne(ctx,
		`SELECT `+aptCols+` FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id
		 WHERE a.id = ?`, id)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Update persists changes to an existing appointment.
func (r *AppointmentRepo) Update(ctx context.Context, a *domain.Appointment) error {
	var freq *string
	if a.FollowUpFreq != nil {
		s := string(*a.FollowUpFreq)
		freq = &s
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET patient_id=?, user_id=?, date=?, time=?, type=?, status=?,
		 notes=?, follow_up_freq=?, updated_at=datetime('now')
		 WHERE id=?`,
		a.PatientID, a.UserID, a.Date.Format("2006-01-02"), a.Time,
		a.Type, a.Status, a.Notes, freq, a.ID,
	)
	return err
}

// Delete removes an appointment.
func (r *AppointmentRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM appointments WHERE id = ?`, id)
	return err
}

// List returns paginated appointments matching the filter.
func (r *AppointmentRepo) List(ctx context.Context, f domain.AppointmentFilter) ([]domain.Appointment, int, error) {
	where, args := buildAptWhere(f)

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM appointments a"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	perPage := f.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	rows, err := r.db.QueryContext(ctx,
		`SELECT `+aptCols+` FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id`+where+
			` ORDER BY a.date, a.time LIMIT ? OFFSET ?`,
		append(args, perPage, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	apts, err := r.scanRows(rows)
	return apts, total, err
}

// ListByDate returns all appointments for a given day.
func (r *AppointmentRepo) ListByDate(ctx context.Context, date time.Time) ([]domain.Appointment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+aptCols+` FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id
		 WHERE a.date = ? ORDER BY a.time`,
		date.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ListByWeek returns all appointments for a 7-day period starting at start.
func (r *AppointmentRepo) ListByWeek(ctx context.Context, start time.Time) ([]domain.Appointment, error) {
	end := start.AddDate(0, 0, 7)
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+aptCols+` FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id
		 WHERE a.date >= ? AND a.date < ? ORDER BY a.date, a.time`,
		start.Format("2006-01-02"), end.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ListOverdue returns confirmed/pending appointments in the past that were never completed.
func (r *AppointmentRepo) ListOverdue(ctx context.Context) ([]domain.Appointment, error) {
	today := time.Now().Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+aptCols+` FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id
		 WHERE a.date < ? AND a.status IN ('confirme','en_attente')
		 ORDER BY a.date`,
		today,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// AvailableSlots returns time slots for a given date with availability status.
func (r *AppointmentRepo) AvailableSlots(ctx context.Context, date time.Time) ([]domain.Slot, error) {
	// Fetch center config for slot generation.
	var startTime, endTime string
	var slotDur int
	err := r.db.QueryRowContext(ctx,
		`SELECT start_time, end_time, slot_duration FROM center WHERE id = 1`,
	).Scan(&startTime, &endTime, &slotDur)
	if err != nil {
		return nil, fmt.Errorf("read center config: %w", err)
	}

	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)

	// Fetch booked slots.
	rows, err := r.db.QueryContext(ctx,
		`SELECT time FROM appointments WHERE date = ? AND status NOT IN ('annule')`,
		date.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	booked := make(map[string]bool)
	for rows.Next() {
		var t string
		rows.Scan(&t)
		booked[t] = true
	}

	var slots []domain.Slot
	for t := start; t.Before(end); t = t.Add(time.Duration(slotDur) * time.Minute) {
		label := t.Format("15:04")
		slots = append(slots, domain.Slot{
			Time:      label,
			Available: !booked[label],
		})
	}
	return slots, nil
}

// CountTodayByStatus returns appointment counts for today grouped by status.
func (r *AppointmentRepo) CountTodayByStatus(ctx context.Context) (map[domain.AppointmentStatus]int, error) {
	today := time.Now().Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM appointments WHERE date = ? GROUP BY status`, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[domain.AppointmentStatus]int)
	for rows.Next() {
		var status domain.AppointmentStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

// --- internal helpers ---

const aptCols = `a.id, a.patient_id, a.user_id, a.date, a.time, a.type, a.status,
	a.notes, a.follow_up_freq, a.created_by, a.created_at, a.updated_at,
	COALESCE(p.last_name || ' ' || p.first_name, ''), COALESCE(p.code, '')`

func (r *AppointmentRepo) scanOne(ctx context.Context, query string, args ...any) (*domain.Appointment, error) {
	a := &domain.Appointment{}
	var dateStr, created, updated string
	var freq *string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&a.ID, &a.PatientID, &a.UserID, &dateStr, &a.Time, &a.Type, &a.Status,
		&a.Notes, &freq, &a.CreatedBy, &created, &updated,
		&a.PatientName, &a.PatientCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAppointmentNotFound
	}
	if err != nil {
		return nil, err
	}
	a.Date = parseTimeVal(dateStr)
	a.CreatedAt = parseTimeVal(created)
	a.UpdatedAt = parseTimeVal(updated)
	if freq != nil {
		f := domain.FollowUpFreq(*freq)
		a.FollowUpFreq = &f
	}
	return a, nil
}

func (r *AppointmentRepo) scanRows(rows *sql.Rows) ([]domain.Appointment, error) {
	var apts []domain.Appointment
	for rows.Next() {
		var a domain.Appointment
		var dateStr, created, updated string
		var freq *string
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.UserID, &dateStr, &a.Time, &a.Type, &a.Status,
			&a.Notes, &freq, &a.CreatedBy, &created, &updated,
			&a.PatientName, &a.PatientCode,
		); err != nil {
			return nil, err
		}
		a.Date = parseTimeVal(dateStr)
		a.CreatedAt = parseTimeVal(created)
		a.UpdatedAt = parseTimeVal(updated)
		if freq != nil {
			f := domain.FollowUpFreq(*freq)
			a.FollowUpFreq = &f
		}
		apts = append(apts, a)
	}
	return apts, rows.Err()
}

func buildAptWhere(f domain.AppointmentFilter) (string, []any) {
	var clauses []string
	var args []any
	if f.PatientID != nil {
		clauses = append(clauses, "a.patient_id = ?")
		args = append(args, *f.PatientID)
	}
	if f.DateFrom != nil {
		clauses = append(clauses, "a.date >= ?")
		args = append(args, f.DateFrom.Format("2006-01-02"))
	}
	if f.DateTo != nil {
		clauses = append(clauses, "a.date <= ?")
		args = append(args, f.DateTo.Format("2006-01-02"))
	}
	if f.Status != nil {
		clauses = append(clauses, "a.status = ?")
		args = append(args, *f.Status)
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	return where, args
}
