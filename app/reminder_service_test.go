package app

import (
	"testing"
	"time"

	"github.com/masante/masante/domain"
)

func mustParseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestRenderTemplate(t *testing.T) {
	p := &domain.Patient{FirstName: "Nathalie", LastName: "Essomba"}
	a := domain.Appointment{Time: "10:00"}
	a.Date = mustParseDate("2026-04-15")
	c := &domain.Center{Name: "Hopital Laquintinie"}

	tpl := "Bonjour {prenom}, RDV le {date} a {heure} au {centre}."
	got := renderTemplate(tpl, p, a, c)
	want := "Bonjour Nathalie, RDV le 15/04/2026 a 10:00 au Hopital Laquintinie."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTemplate_MissingCenter(t *testing.T) {
	p := &domain.Patient{FirstName: "Paul"}
	a := domain.Appointment{Time: "08:00"}
	a.Date = mustParseDate("2026-01-01")

	got := renderTemplate("Bonjour {prenom}, {centre}", p, a, nil)
	if got != "Bonjour Paul, " {
		t.Errorf("got %q", got)
	}
}

func TestRenderTemplate_AllPlaceholders(t *testing.T) {
	p := &domain.Patient{FirstName: "Marie", LastName: "Ngassa"}
	a := domain.Appointment{Time: "14:30"}
	a.Date = mustParseDate("2026-12-25")
	c := &domain.Center{Name: "Centre de Sante Akwa"}

	tpl := "{prenom} {nom} — {date} {heure} — {centre}"
	got := renderTemplate(tpl, p, a, c)
	want := "Marie Ngassa — 25/12/2026 14:30 — Centre de Sante Akwa"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
