package domain

import "testing"

func TestUser_CanAccess(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		resource string
		want     bool
	}{
		{"admin accesses anything", RoleAdmin, "patient", true},
		{"admin accesses unknown", RoleAdmin, "unknown_resource", true},
		{"medecin accesses patient", RoleMedecin, "patient", true},
		{"medecin accesses dashboard", RoleMedecin, "dashboard", true},
		{"medecin cannot manage users", RoleMedecin, "user.manage", false},
		{"infirmier accesses reminder", RoleInfirmier, "reminder", true},
		{"infirmier cannot export", RoleInfirmier, "export", false},
		{"asc accesses own module", RoleASC, "asc", true},
		{"asc reads patients", RoleASC, "patient.read", true},
		{"asc cannot write patients", RoleASC, "patient", false},
		{"asc cannot access dashboard", RoleASC, "dashboard", false},
		{"unknown role denied", Role("unknown"), "patient", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.CanAccess(tt.resource); got != tt.want {
				t.Errorf("CanAccess(%q) = %v, want %v", tt.resource, got, tt.want)
			}
		})
	}
}
