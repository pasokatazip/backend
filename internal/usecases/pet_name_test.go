package usecases

import (
	"strings"
	"testing"
)

func TestNormalizeAndValidatePetName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantValid bool
	}{
		{name: "accepts one character", input: "ポ", want: "ポ", wantValid: true},
		{name: "accepts internal spaces", input: "ポチ 太郎", want: "ポチ 太郎", wantValid: true},
		{name: "trims surrounding spaces", input: "  ポチ 太郎  ", want: "ポチ 太郎", wantValid: true},
		{name: "accepts thirty characters", input: strings.Repeat("あ", 30), want: strings.Repeat("あ", 30), wantValid: true},
		{name: "rejects an empty name", input: ""},
		{name: "rejects an ASCII whitespace-only name", input: "   "},
		{name: "rejects a Unicode whitespace-only name", input: "　　"},
		{name: "rejects thirty-one characters", input: strings.Repeat("あ", 31)},
		{name: "rejects a newline", input: "ポチ\n太郎"},
		{name: "rejects a carriage return", input: "ポチ\r太郎"},
		{name: "rejects a tab", input: "ポチ\t太郎"},
		{name: "rejects a leading control character before trimming", input: "\nポチ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := normalizeAndValidatePetName(tt.input)
			if valid != tt.wantValid || got != tt.want {
				t.Fatalf("normalizeAndValidatePetName() = (%q, %v), want (%q, %v)", got, valid, tt.want, tt.wantValid)
			}
		})
	}
}
