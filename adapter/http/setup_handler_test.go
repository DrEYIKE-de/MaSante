package http

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masante/masante/adapter"
	"github.com/masante/masante/adapter/sqlite"
	"github.com/masante/masante/app"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	db, err := sqlite.Open(dir + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.Migrate(db); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	userRepo := sqlite.NewUserRepo(db)
	sessionRepo := sqlite.NewSessionRepo(db)
	centerRepo := sqlite.NewCenterRepo(db)
	smsRepo := sqlite.NewSMSConfigRepo(db)
	auditRepo := sqlite.NewAuditRepo(db)
	hasher := adapter.BcryptHasher{}

	patientRepo := sqlite.NewPatientRepo(db)
	aptRepo := sqlite.NewAppointmentRepo(db)

	reminderRepo := sqlite.NewReminderRepo(db)

	authSvc := app.NewAuthService(userRepo, sessionRepo, hasher, auditRepo)
	setupSvc := app.NewSetupService(centerRepo, userRepo, smsRepo, hasher, auditRepo)
	patientSvc := app.NewPatientService(patientRepo, auditRepo)
	aptSvc := app.NewAppointmentService(aptRepo, patientRepo, auditRepo)
	userSvc := app.NewUserService(userRepo, sessionRepo, hasher, auditRepo)
	reminderSvc := app.NewReminderService(reminderRepo, aptRepo, patientRepo, smsRepo, centerRepo)

	return NewServer(authSvc, setupSvc, patientSvc, aptSvc, userSvc, reminderSvc)
}

func TestSetupStatus_InitiallyFalse(t *testing.T) {
	srv := testServer(t)

	req := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"setup_complete":false`) {
		t.Errorf("body = %s, want setup_complete false", w.Body.String())
	}
}

func TestSetupCenter_OK(t *testing.T) {
	srv := testServer(t)

	body := `{"name":"Hopital Test","type":"hopital_public","country":"Cameroun","city":"Douala"}`
	req := httptest.NewRequest("POST", "/api/v1/setup/center", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestSetupCenter_MissingFields(t *testing.T) {
	srv := testServer(t)

	body := `{"name":"Test"}`
	req := httptest.NewRequest("POST", "/api/v1/setup/center", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSetupAdmin_TooShortPassword(t *testing.T) {
	srv := testServer(t)

	body := `{"full_name":"Admin","username":"admin","password":"short"}`
	req := httptest.NewRequest("POST", "/api/v1/setup/admin", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSetup_BlockedAfterComplete(t *testing.T) {
	srv := testServer(t)

	// Do the full setup
	steps := []struct {
		path string
		body string
	}{
		{"/api/v1/setup/center", `{"name":"Test","type":"centre_sante","country":"Cameroun","city":"Douala"}`},
		{"/api/v1/setup/admin", `{"full_name":"Admin","username":"admin","password":"longpassword1","email":"a@b.cm"}`},
		{"/api/v1/setup/schedule", `{"consultation_days":"1,2,3,4,5","start_time":"08:00","end_time":"16:00","slot_duration":30,"max_patients_day":40}`},
		{"/api/v1/setup/sms", `{"enabled":false}`},
		{"/api/v1/setup/complete", `{}`},
	}
	for _, s := range steps {
		req := httptest.NewRequest("POST", s.path, strings.NewReader(s.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("%s: status = %d, body = %s", s.path, w.Code, w.Body.String())
		}
	}

	// Now setup should be blocked
	req := httptest.NewRequest("POST", "/api/v1/setup/center", strings.NewReader(`{"name":"Hack","type":"centre_sante","country":"X","city":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("status = %d, want 403 after setup complete", w.Code)
	}
}
