package validation

import "testing"

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		wantErr bool
	}{
		{"too short", "Ab1!xxxxxxx", true},
		{"12 chars valid", "Abcdef12!@#$", false},
		{"no uppercase", "abcdef12!@#$", true},
		{"no lowercase", "ABCDEF12!@#$", true},
		{"no digit", "Abcdefgh!@#$", true},
		{"no special", "Abcdefgh1234", true},
		{"exactly 12", "Abcd1234!@#$", false},
		{"11 chars", "Abc123!@#$x", true},
		{"unicode letters not special", "Abcdefgh123é", true},
		{"emoji as special", "Abcdef1234!😀", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pass)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) = %v, wantErr %v", tt.pass, err, tt.wantErr)
			}
		})
	}
}
