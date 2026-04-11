package sqlite

import (
	"database/sql"
	"fmt"
)

var migrations = []string{
	// v1: schema initial
	`CREATE TABLE IF NOT EXISTS center (
		id              INTEGER PRIMARY KEY CHECK (id = 1),
		name            TEXT NOT NULL,
		type            TEXT NOT NULL CHECK (type IN ('hopital_public','centre_sante','clinique_privee')),
		country         TEXT NOT NULL,
		city            TEXT NOT NULL,
		district        TEXT NOT NULL DEFAULT '',
		latitude        REAL,
		longitude       REAL,
		consultation_days TEXT NOT NULL DEFAULT '1,2,3,4,5',
		start_time      TEXT NOT NULL DEFAULT '08:00',
		end_time        TEXT NOT NULL DEFAULT '16:00',
		slot_duration   INTEGER NOT NULL DEFAULT 30,
		max_patients_day INTEGER NOT NULL DEFAULT 40,
		setup_step      INTEGER NOT NULL DEFAULT 0,
		setup_complete  INTEGER NOT NULL DEFAULT 0,
		created_at      TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS users (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		username        TEXT NOT NULL UNIQUE COLLATE NOCASE,
		password_hash   TEXT NOT NULL,
		full_name       TEXT NOT NULL,
		email           TEXT NOT NULL DEFAULT '',
		phone           TEXT NOT NULL DEFAULT '',
		role            TEXT NOT NULL CHECK (role IN ('admin','medecin','infirmier','asc')),
		title           TEXT NOT NULL DEFAULT '',
		status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','conge','desactive')),
		must_change_pwd INTEGER NOT NULL DEFAULT 0,
		last_login_at   TEXT,
		created_at      TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token       TEXT PRIMARY KEY,
		user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at  TEXT NOT NULL DEFAULT (datetime('now')),
		expires_at  TEXT NOT NULL,
		ip_address  TEXT NOT NULL DEFAULT '',
		user_agent  TEXT NOT NULL DEFAULT ''
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

	CREATE TABLE IF NOT EXISTS patients (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		code            TEXT NOT NULL UNIQUE,
		last_name       TEXT NOT NULL,
		first_name      TEXT NOT NULL,
		date_of_birth   TEXT,
		sex             TEXT NOT NULL CHECK (sex IN ('M','F')),
		phone           TEXT NOT NULL DEFAULT '',
		phone_secondary TEXT NOT NULL DEFAULT '',
		district        TEXT NOT NULL DEFAULT '',
		address         TEXT NOT NULL DEFAULT '',
		language        TEXT NOT NULL DEFAULT 'fr',
		reminder_channel TEXT NOT NULL DEFAULT 'sms' CHECK (reminder_channel IN ('sms','whatsapp','voice','none')),
		contact_name    TEXT NOT NULL DEFAULT '',
		contact_phone   TEXT NOT NULL DEFAULT '',
		contact_relation TEXT NOT NULL DEFAULT '',
		referred_by     TEXT NOT NULL DEFAULT '',
		status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','a_surveiller','perdu_de_vue','sorti')),
		risk_score      INTEGER NOT NULL DEFAULT 5 CHECK (risk_score BETWEEN 0 AND 10),
		enrollment_date TEXT NOT NULL DEFAULT (date('now')),
		exit_reason     TEXT CHECK (exit_reason IN ('deces','transfert','abandon','perdu_de_vue','guerison',NULL)),
		exit_date       TEXT,
		exit_notes      TEXT NOT NULL DEFAULT '',
		created_at      TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_patients_code ON patients(code);
	CREATE INDEX IF NOT EXISTS idx_patients_status ON patients(status);
	CREATE INDEX IF NOT EXISTS idx_patients_name ON patients(last_name, first_name);

	CREATE TABLE IF NOT EXISTS appointments (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id      INTEGER NOT NULL REFERENCES patients(id),
		user_id         INTEGER REFERENCES users(id),
		date            TEXT NOT NULL,
		time            TEXT NOT NULL,
		type            TEXT NOT NULL CHECK (type IN ('consultation','retrait_medicaments','bilan_sanguin','club_adherence')),
		status          TEXT NOT NULL DEFAULT 'confirme' CHECK (status IN ('confirme','en_attente','termine','manque','annule','reporte')),
		notes           TEXT NOT NULL DEFAULT '',
		follow_up_freq  TEXT CHECK (follow_up_freq IN ('mensuel','trimestriel','semestriel',NULL)),
		created_by      INTEGER REFERENCES users(id),
		created_at      TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_appointments_patient ON appointments(patient_id);
	CREATE INDEX IF NOT EXISTS idx_appointments_date ON appointments(date);
	CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);

	CREATE TABLE IF NOT EXISTS reminders (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		appointment_id  INTEGER NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
		patient_id      INTEGER NOT NULL REFERENCES patients(id),
		channel         TEXT NOT NULL CHECK (channel IN ('sms','whatsapp','voice')),
		type            TEXT NOT NULL CHECK (type IN ('j7','j2','j0','retard')),
		message         TEXT NOT NULL,
		status          TEXT NOT NULL DEFAULT 'planifie' CHECK (status IN ('planifie','envoye','recu','echec')),
		scheduled_at    TEXT NOT NULL,
		sent_at         TEXT,
		provider_id     TEXT NOT NULL DEFAULT '',
		error_message   TEXT NOT NULL DEFAULT '',
		retry_count     INTEGER NOT NULL DEFAULT 0,
		created_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_reminders_status ON reminders(status);
	CREATE INDEX IF NOT EXISTS idx_reminders_scheduled ON reminders(scheduled_at);

	CREATE TABLE IF NOT EXISTS message_templates (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		name      TEXT NOT NULL UNIQUE,
		channel   TEXT NOT NULL DEFAULT 'sms',
		body      TEXT NOT NULL,
		language  TEXT NOT NULL DEFAULT 'fr',
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS asc_visits (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id      INTEGER NOT NULL REFERENCES patients(id),
		asc_user_id     INTEGER NOT NULL REFERENCES users(id),
		visit_date      TEXT NOT NULL DEFAULT (date('now')),
		patient_found   TEXT NOT NULL CHECK (patient_found IN ('oui','non','demenage')),
		absence_reason  TEXT NOT NULL DEFAULT '',
		notes           TEXT NOT NULL DEFAULT '',
		next_appointment_id INTEGER REFERENCES appointments(id),
		created_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_asc_visits_patient ON asc_visits(patient_id);

	CREATE TABLE IF NOT EXISTS sms_config (
		id              INTEGER PRIMARY KEY CHECK (id = 1),
		enabled         INTEGER NOT NULL DEFAULT 0,
		provider        TEXT NOT NULL DEFAULT '' CHECK (provider IN ('','africastalking','twilio','orange','mtn','infobip')),
		api_key         TEXT NOT NULL DEFAULT '',
		api_secret      TEXT NOT NULL DEFAULT '',
		sender_id       TEXT NOT NULL DEFAULT '',
		reminder_j7     INTEGER NOT NULL DEFAULT 1,
		reminder_j2     INTEGER NOT NULL DEFAULT 1,
		reminder_j0     INTEGER NOT NULL DEFAULT 0,
		reminder_late   INTEGER NOT NULL DEFAULT 1,
		late_delay_days INTEGER NOT NULL DEFAULT 3,
		updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS settings (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS audit_log (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER REFERENCES users(id),
		action      TEXT NOT NULL,
		entity_type TEXT NOT NULL DEFAULT '',
		entity_id   INTEGER,
		details     TEXT NOT NULL DEFAULT '',
		ip_address  TEXT NOT NULL DEFAULT '',
		created_at  TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);

	-- Default message templates
	INSERT OR IGNORE INTO message_templates (name, body, language)
	VALUES
		('rappel_j7', 'Bonjour {prenom}, votre prochain rendez-vous sante est prevu le {date} a {heure}. Pensez a apporter votre carnet. A bientot!', 'fr'),
		('rappel_j2', 'Bonjour {prenom}, ceci est un rappel de votre rendez-vous sante prevu le {date} a {heure} au centre de sante. Merci de confirmer en repondant OUI.', 'fr'),
		('rappel_j0', 'Bonjour {prenom}, votre rendez-vous sante est aujourd''hui a {heure}. Nous vous attendons.', 'fr'),
		('retard', 'Bonjour {prenom}, vous avez un rendez-vous sante en attente. Merci de contacter le centre pour reprogrammer.', 'fr');`,

	// v2: account lockout fields
	`ALTER TABLE users ADD COLUMN failed_attempts INTEGER NOT NULL DEFAULT 0;
	 ALTER TABLE users ADD COLUMN locked_until TEXT;`,

	// v3: missing indexes
	`CREATE INDEX IF NOT EXISTS idx_reminders_appointment ON reminders(appointment_id);
	 CREATE INDEX IF NOT EXISTS idx_reminders_patient ON reminders(patient_id);
	 CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id);`,
}

func Migrate(db *DB) error {
	_, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	var current int
	err = db.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&current)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	for i := current; i < len(migrations); i++ {
		if err := applyMigration(db.conn, i+1, migrations[i]); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}

	return nil
}

func applyMigration(conn *sql.DB, version int, sql string) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(sql); err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
		return fmt.Errorf("record: %w", err)
	}

	return tx.Commit()
}
