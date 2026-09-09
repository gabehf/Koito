package cfg

import "testing"

func TestIsValidDateFormat(t *testing.T) {
	valid := []string{
		"DD/MM/YYYY",
		"MM-DD-YYYY",
		"YYYY.MM.DD",
		"MM/DD/YYYY",
	}
	for _, s := range valid {
		if !isValidDateFormat(s) {
			t.Errorf("isValidDateFormat(%q) = false, want true", s)
		}
	}

	invalid := []string{
		"DD-MM/YYYY",    // mixed separators
		"DD/DD/YYYY",    // duplicate token
		"DD/MM",         // missing token
		"DD/MM/YYYY/DD", // extra token
		"dd/mm/yyyy",    // wrong case
		"",
	}
	for _, s := range invalid {
		if isValidDateFormat(s) {
			t.Errorf("isValidDateFormat(%q) = true, want false", s)
		}
	}
}
