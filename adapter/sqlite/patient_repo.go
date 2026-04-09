package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/masante/masante/domain"
)

// PatientRepo implements domain.PatientRepository backed by SQLite.
type PatientRepo struct {
	db *sql.DB
}

// NewPatientRepo returns a new PatientRepo.
func NewPatientRepo(db *DB) *PatientRepo {
	return &PatientRepo{db: db.conn}
}

// Create inserts a new patient. It sets p.ID on success.
func (r *PatientRepo) Create(ctx context.Context, p *domain.Patient) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO patients
		 (code, last_name, first_name, date_of_birth, sex, phone, phone_secondary,
		  district, address, language, reminder_channel, contact_name, contact_phone,
		  contact_relation, referred_by, status, risk_score, enrollment_date)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.Code, p.LastName, p.FirstName, formatDatePtr(p.DateOfBirth), p.Sex,
		p.Phone, p.PhoneSecondary, p.District, p.Address, p.Language,
		p.ReminderChannel, p.ContactName, p.ContactPhone, p.ContactRelation,
		p.ReferredBy, p.Status, p.RiskScore, p.EnrollmentDate.Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("insert patient: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

// GetByID returns a patient by primary key.
func (r *PatientRepo) GetByID(ctx context.Context, id int64) (*domain.Patient, error) {
	return r.scanOne(ctx, `SELECT `+patientCols+` FROM patients WHERE id = ?`, id)
}

// GetByCode returns a patient by unique code.
func (r *PatientRepo) GetByCode(ctx context.Context, code string) (*domain.Patient, error) {
	return r.scanOne(ctx, `SELECT `+patientCols+` FROM patients WHERE code = ?`, code)
}

// Update persists changes to an existing patient.
func (r *PatientRepo) Update(ctx context.Context, p *domain.Patient) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE patients SET
		 last_name=?, first_name=?, date_of_birth=?, sex=?, phone=?, phone_secondary=?,
		 district=?, address=?, language=?, reminder_channel=?, contact_name=?, contact_phone=?,
		 contact_relation=?, referred_by=?, status=?, risk_score=?,
		 exit_reason=?, exit_date=?, exit_notes=?, updated_at=datetime('now')
		 WHERE id=?`,
		p.LastName, p.FirstName, formatDatePtr(p.DateOfBirth), p.Sex,
		p.Phone, p.PhoneSecondary, p.District, p.Address, p.Language,
		p.ReminderChannel, p.ContactName, p.ContactPhone, p.ContactRelation,
		p.ReferredBy, p.Status, p.RiskScore,
		p.ExitReason, formatDatePtr(p.ExitDate), p.ExitNotes, p.ID,
	)
	return err
}

// List returns paginated patients matching the filter.
// The second return value is the total count before pagination.
func (r *PatientRepo) List(ctx context.Context, f domain.PatientFilter) ([]domain.Patient, int, error) {
	where, args := buildPatientWhere(f)

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM patients"+where, args...).Scan(&total); err != nil {
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
		`SELECT `+patientCols+` FROM patients`+where+` ORDER BY last_name, first_name LIMIT ? OFFSET ?`,
		append(args, perPage, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	patients, err := r.scanRows(rows)
	return patients, total, err
}

// Search performs a free-text search on name, code, and phone.
func (r *PatientRepo) Search(ctx context.Context, query string, limit int) ([]domain.Patient, error) {
	if limit <= 0 {
		limit = 10
	}
	like := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+patientCols+` FROM patients
		 WHERE last_name LIKE ? OR first_name LIKE ? OR code LIKE ? OR phone LIKE ?
		 ORDER BY last_name, first_name LIMIT ?`,
		like, like, like, like, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// NextCode generates the next unique patient code in format MS-YYYY-NNNNN.
func (r *PatientRepo) NextCode(ctx context.Context) (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("MS-%d-", year)

	var maxCode sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(code) FROM patients WHERE code LIKE ?`, prefix+"%",
	).Scan(&maxCode)
	if err != nil {
		return "", err
	}

	seq := 1
	if maxCode.Valid && len(maxCode.String) > len(prefix) {
		fmt.Sscanf(maxCode.String[len(prefix):], "%d", &seq)
		seq++
	}

	return fmt.Sprintf("MS-%d-%05d", year, seq), nil
}

// CountByStatus returns patient counts grouped by status.
func (r *PatientRepo) CountByStatus(ctx context.Context) (map[domain.PatientStatus]int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM patients GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[domain.PatientStatus]int)
	for rows.Next() {
		var status domain.PatientStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

// --- internal helpers ---

const patientCols = `id, code, last_name, first_name, date_of_birth, sex, phone, phone_secondary,
	district, address, language, reminder_channel, contact_name, contact_phone,
	contact_relation, referred_by, status, risk_score, enrollment_date,
	exit_reason, exit_date, exit_notes, created_at, updated_at`

func (r *PatientRepo) scanOne(ctx context.Context, query string, args ...any) (*domain.Patient, error) {
	p := &domain.Patient{}
	var dob, exitDate, enrollment, created, updated *string
	var exitReason *string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Code, &p.LastName, &p.FirstName, &dob, &p.Sex,
		&p.Phone, &p.PhoneSecondary, &p.District, &p.Address, &p.Language,
		&p.ReminderChannel, &p.ContactName, &p.ContactPhone, &p.ContactRelation,
		&p.ReferredBy, &p.Status, &p.RiskScore, &enrollment,
		&exitReason, &exitDate, &p.ExitNotes, &created, &updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPatientNotFound
	}
	if err != nil {
		return nil, err
	}
	p.DateOfBirth = parseTime(dob)
	p.ExitDate = parseTime(exitDate)
	if enrollment != nil {
		p.EnrollmentDate = parseTimeVal(*enrollment)
	}
	if created != nil {
		p.CreatedAt = parseTimeVal(*created)
	}
	if updated != nil {
		p.UpdatedAt = parseTimeVal(*updated)
	}
	if exitReason != nil {
		r := domain.ExitReason(*exitReason)
		p.ExitReason = &r
	}
	return p, nil
}

func (r *PatientRepo) scanRows(rows *sql.Rows) ([]domain.Patient, error) {
	var patients []domain.Patient
	for rows.Next() {
		var p domain.Patient
		var dob, exitDate, enrollment, created, updated *string
		var exitReason *string
		if err := rows.Scan(
			&p.ID, &p.Code, &p.LastName, &p.FirstName, &dob, &p.Sex,
			&p.Phone, &p.PhoneSecondary, &p.District, &p.Address, &p.Language,
			&p.ReminderChannel, &p.ContactName, &p.ContactPhone, &p.ContactRelation,
			&p.ReferredBy, &p.Status, &p.RiskScore, &enrollment,
			&exitReason, &exitDate, &p.ExitNotes, &created, &updated,
		); err != nil {
			return nil, err
		}
		p.DateOfBirth = parseTime(dob)
		p.ExitDate = parseTime(exitDate)
		if enrollment != nil {
			p.EnrollmentDate = parseTimeVal(*enrollment)
		}
		if created != nil {
			p.CreatedAt = parseTimeVal(*created)
		}
		if updated != nil {
			p.UpdatedAt = parseTimeVal(*updated)
		}
		if exitReason != nil {
			r := domain.ExitReason(*exitReason)
			p.ExitReason = &r
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func buildPatientWhere(f domain.PatientFilter) (string, []any) {
	where := " WHERE 1=1"
	var args []any
	if f.Status != nil {
		where += " AND status = ?"
		args = append(args, *f.Status)
	}
	if f.District != "" {
		where += " AND district = ?"
		args = append(args, f.District)
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		where += " AND (last_name LIKE ? OR first_name LIKE ? OR code LIKE ?)"
		args = append(args, like, like, like)
	}
	return where, args
}

func formatDatePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02")
}
