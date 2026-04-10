package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/masante/masante/domain"
)

type CenterRepo struct {
	db *sql.DB
}

func NewCenterRepo(db *DB) *CenterRepo {
	return &CenterRepo{db: db.conn}
}

func (r *CenterRepo) Get(ctx context.Context) (*domain.Center, error) {
	c := &domain.Center{}
	var created, updated string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, type, country, city, district, latitude, longitude,
		        consultation_days, start_time, end_time, slot_duration, max_patients_day,
		        setup_step, setup_complete, created_at, updated_at
		 FROM center WHERE id = 1`,
	).Scan(&c.ID, &c.Name, &c.Type, &c.Country, &c.City, &c.District,
		&c.Latitude, &c.Longitude, &c.ConsultationDays, &c.StartTime, &c.EndTime,
		&c.SlotDuration, &c.MaxPatientsDay, &c.SetupStep, &c.SetupComplete, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.CreatedAt = parseTimeVal(created)
	c.UpdatedAt = parseTimeVal(updated)
	return c, nil
}

func (r *CenterRepo) Create(ctx context.Context, c *domain.Center) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO center (id, name, type, country, city, district, latitude, longitude)
		 VALUES (1, ?, ?, ?, ?, ?, ?, ?)`,
		c.Name, c.Type, c.Country, c.City, c.District, c.Latitude, c.Longitude,
	)
	return err
}

func (r *CenterRepo) Update(ctx context.Context, c *domain.Center) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE center SET name=?, type=?, country=?, city=?, district=?, latitude=?, longitude=?,
		        consultation_days=?, start_time=?, end_time=?, slot_duration=?, max_patients_day=?,
		        updated_at=datetime('now')
		 WHERE id = 1`,
		c.Name, c.Type, c.Country, c.City, c.District, c.Latitude, c.Longitude,
		c.ConsultationDays, c.StartTime, c.EndTime, c.SlotDuration, c.MaxPatientsDay,
	)
	return err
}

func (r *CenterRepo) SetSetupStep(ctx context.Context, step int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE center SET setup_step = ?, updated_at = datetime('now') WHERE id = 1`, step)
	return err
}

func (r *CenterRepo) CompleteSetup(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE center SET setup_step = 5, setup_complete = 1, updated_at = datetime('now') WHERE id = 1`)
	return err
}

func (r *CenterRepo) GetSetupStep(ctx context.Context) (int, error) {
	var step int
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(setup_step, 0) FROM center WHERE id = 1`).Scan(&step)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return step, err
}

func (r *CenterRepo) IsSetupDone(ctx context.Context) (bool, error) {
	var done bool
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(setup_complete, 0) FROM center WHERE id = 1`).Scan(&done)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return done, err
}
