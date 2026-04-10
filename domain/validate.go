package domain

import "errors"

// ErrInvalidValue is returned when a field value is not in the allowed set.
var ErrInvalidValue = errors.New("valeur invalide")

// ValidateRole checks that r is a known role.
func ValidateRole(r Role) error {
	switch r {
	case RoleAdmin, RoleMedecin, RoleInfirmier, RoleASC:
		return nil
	}
	return ErrInvalidValue
}

// ValidateUserStatus checks that s is a known user status.
func ValidateUserStatus(s UserStatus) error {
	switch s {
	case UserActive, UserOnLeave, UserDisabled:
		return nil
	}
	return ErrInvalidValue
}

// ValidateReminderChannel checks that c is a known channel.
func ValidateReminderChannel(c ReminderChannel) error {
	switch c {
	case ChannelSMS, ChannelWhatsApp, ChannelVoice, ChannelNone:
		return nil
	}
	return ErrInvalidValue
}

// ValidateAppointmentType checks that t is a known appointment type.
func ValidateAppointmentType(t AppointmentType) error {
	switch t {
	case TypeConsultation, TypeMedPickup, TypeBloodTest, TypeAdherence:
		return nil
	}
	return ErrInvalidValue
}

// ValidateExitReason checks that r is a known exit reason.
func ValidateExitReason(r ExitReason) error {
	switch r {
	case ExitDeath, ExitTransfer, ExitDropout, ExitLost, ExitCured:
		return nil
	}
	return ErrInvalidValue
}

// ValidateSex checks that s is M or F.
func ValidateSex(s string) error {
	if s == "M" || s == "F" {
		return nil
	}
	return ErrInvalidValue
}

// ValidateTimeFormat checks that t is a valid HH:MM string.
func ValidateTimeFormat(t string) error {
	if len(t) != 5 || t[2] != ':' {
		return ErrInvalidValue
	}
	h := (t[0]-'0')*10 + (t[1] - '0')
	m := (t[3]-'0')*10 + (t[4] - '0')
	if h > 23 || m > 59 {
		return ErrInvalidValue
	}
	return nil
}
