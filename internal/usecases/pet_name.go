package usecases

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxPetNameLength = 30

func normalizeAndValidatePetName(value string) (string, bool) {
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", false
		}
	}

	normalized := strings.TrimSpace(value)
	length := utf8.RuneCountInString(normalized)
	if length < 1 || length > maxPetNameLength {
		return "", false
	}

	return normalized, true
}
