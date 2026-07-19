package validation

import (
	"strings"
	"unicode"
)

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)

	if len(email) == 0 || len(email) > 254 {
		return false
	}

	for _, r := range email {
		if unicode.IsControl(r) {
			return false
		}
	}

	atIdx := strings.LastIndex(email, "@")
	if atIdx < 1 {
		return false
	}

	local := email[:atIdx]
	domain := email[atIdx+1:]

	if len(local) == 0 || len(local) > 64 {
		return false
	}
	if len(domain) == 0 {
		return false
	}

	if !isValidLocal(local) {
		return false
	}

	if !isValidDomain(domain) {
		return false
	}

	return true
}

func isValidLocal(local string) bool {
	if local[0] == '.' || local[len(local)-1] == '.' {
		return false
	}

	for i := 0; i < len(local); i++ {
		c := local[i]
		if !(c >= 'a' && c <= 'z') &&
			!(c >= 'A' && c <= 'Z') &&
			!(c >= '0' && c <= '9') &&
			c != '.' && c != '_' && c != '+' && c != '-' {
			return false
		}
	}

	return true
}

func isValidDomain(domain string) bool {
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}

	for _, label := range labels {
		if len(label) == 0 {
			return false
		}
		if len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z') &&
				!(c >= 'A' && c <= 'Z') &&
				!(c >= '0' && c <= '9') &&
				c != '-' {
				return false
			}
		}
	}

	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	if strings.HasPrefix(tld, "xn--") && len(tld) < 5 {
		return false
	}

	return true
}
