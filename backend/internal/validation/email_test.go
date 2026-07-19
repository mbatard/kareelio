package validation

import (
	"strings"
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"john.doe@example.fr", true},
		{"john+tag@example.co.uk", true},
		{"user_name@example.travel", true},
		{"test@sub.example.com", true},
		{"Test@Example.COM", true},
		{"a@b.co", true},
		{"user123@example.com", true},
		{"user-name@example.com", true},
		{"user_name@example.com", true},
		{"user.name@example.com", true},

		{"", false},
		{"test", false},
		{"test@", false},
		{"@example.com", false},
		{"test@example", false},
		{"test@example.", false},
		{"test@.example.com", false},
		{"test@example..com", false},
		{"test@-example.com", false},
		{"test@example-.com", false},
		{"test example@example.com", false},
		{"test@example.c", false},
		{".test@example.com", false},
		{"test.@example.com", false},
		{"test@@example.com", false},
		{"a..b@example.com", true},
		{"test@exam ple.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			if got != tt.valid {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Test@Example.COM", "test@example.com"},
		{"  test@example.com  ", "test@example.com"},
		{"Test@Example.com", "test@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeEmail(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidEmailLongLocal(t *testing.T) {
	local := strings.Repeat("a", 65)
	email := local + "@example.com"
	if IsValidEmail(email) {
		t.Errorf("IsValidEmail should reject local part > 64 chars")
	}
}
