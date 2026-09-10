package usecases

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	maxEmailLength        = 255
	minPasswordLength     = 8
	maxPasswordByteLength = 72
)

func normalizeAndValidateEmail(value string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" || utf8.RuneCountInString(normalized) > maxEmailLength {
		return "", false
	}

	address, err := mail.ParseAddress(normalized)
	if err != nil || address.Address != normalized {
		return "", false
	}

	return normalized, true
}

func isValidPassword(value string) bool {
	return strings.TrimSpace(value) != "" &&
		utf8.RuneCountInString(value) >= minPasswordLength &&
		len([]byte(value)) <= maxPasswordByteLength
}
