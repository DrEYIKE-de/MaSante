package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

func TestPatientsToPDF(t *testing.T) {
	patients := []domain.Patient{
		{Code: "MS-2026-00001", LastName: "Essomba", FirstName: "Nathalie", Sex: "F", Phone: "+237600000000", District: "Akwa", Status: domain.PatientActive, RiskScore: 3, EnrollmentDate: time.Now()},
	}

	var buf bytes.Buffer
	if err := PatientsToPDF(&buf, patients, "Patients actifs"); err != nil {
		t.Fatalf("PatientsToPDF: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
	// PDF starts with %PDF.
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("output does not look like a valid PDF")
	}
}

func TestMonthlyReportPDF(t *testing.T) {
	patientCounts := map[domain.PatientStatus]int{
		domain.PatientActive: 100,
	}
	aptCounts := map[domain.AppointmentStatus]int{
		domain.StatusConfirmed: 30,
	}

	var buf bytes.Buffer
	if err := MonthlyReportPDF(&buf, "Mars 2026", "Hopital Laquintinie", patientCounts, aptCounts); err != nil {
		t.Fatalf("MonthlyReportPDF: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("output does not look like a valid PDF")
	}
}
