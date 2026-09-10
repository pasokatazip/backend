package usecases

import (
	"strings"
	"testing"
)

func TestNormalizeAndValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantValid bool
	}{
		{
			name:      "trims surrounding spaces and lowercases",
			input:     "  User.Name+tag@EXAMPLE.COM  ",
			want:      "user.name+tag@example.com",
			wantValid: true,
		},
		{name: "rejects an empty email", input: "   "},
		{name: "rejects a malformed email", input: "not-an-email"},
		{name: "rejects a display name", input: "User <user@example.com>"},
		{
			name:  "rejects more than 255 characters",
			input: strings.Repeat("a", 244) + "@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := normalizeAndValidateEmail(tt.input)
			if valid != tt.wantValid || got != tt.want {
				t.Fatalf("normalizeAndValidateEmail() = (%q, %v), want (%q, %v)", got, valid, tt.want, tt.wantValid)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "accepts eight characters", value: "password", want: true},
		{name: "accepts spaces in a non-blank password", value: "pass word", want: true},
		{name: "rejects whitespace only", value: "        "},
		{name: "rejects fewer than eight characters", value: "pass123"},
		{name: "accepts exactly bcrypt byte limit", value: strings.Repeat("a", 72), want: true},
		{name: "rejects more than bcrypt byte limit", value: strings.Repeat("a", 73)},
		{name: "counts multibyte password bytes for bcrypt limit", value: strings.Repeat("あ", 25)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPassword(tt.value); got != tt.want {
				t.Fatalf("isValidPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
