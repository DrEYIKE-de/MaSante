package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

func TestPatientsToExcel(t *testing.T) {
	patients := []domain.Patient{
		{Code: "MS-2026-00001", LastName: "Essomba", FirstName: "Nathalie", Sex: "F", Phone: "+237600000000", District: "Akwa", Status: domain.PatientActive, RiskScore: 3, EnrollmentDate: time.Now()},
		{Code: "MS-2026-00002", LastName: "Mbouda", FirstName: "Thierry", Sex: "M", Phone: "+237600000001", District: "Bali", Status: domain.PatientMonitored, RiskScore: 7, EnrollmentDate: time.Now()},
	}

	var buf bytes.Buffer
	if err := PatientsToExcel(&buf, patients, "Test"); err != nil {
		t.Fatalf("PatientsToExcel: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
	// XLSX files start with PK (zip).
	if buf.Bytes()[0] != 'P' || buf.Bytes()[1] != 'K' {
		t.Error("output does not look like a valid XLSX file")
	}
}

func TestAppointmentsToExcel(t *testing.T) {
	apts := []domain.Appointment{
		{Date: time.Now(), Time: "08:00", PatientName: "Essomba Nathalie", PatientCode: "MS-2026-00001", Type: domain.TypeConsultation, Status: domain.StatusConfirmed},
	}

	var buf bytes.Buffer
	if err := AppointmentsToExcel(&buf, apts); err != nil {
		t.Fatalf("AppointmentsToExcel: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
}

func TestMonthlyReportExcel(t *testing.T) {
	patientCounts := map[domain.PatientStatus]int{
		domain.PatientActive: 100,
		domain.PatientLost:   5,
	}
	aptCounts := map[domain.AppointmentStatus]int{
		domain.StatusConfirmed: 30,
		domain.StatusMissed:    3,
	}

	var buf bytes.Buffer
	if err := MonthlyReportExcel(&buf, "Mars 2026", patientCounts, aptCounts); err != nil {
		t.Fatalf("MonthlyReportExcel: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
}
