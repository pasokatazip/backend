package persistence

import (
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

func TestCalculateAwardedExperience(t *testing.T) {
	tests := []struct {
		name        string
		amount      int
		usages      []experienceCapUsage
		wantAwarded int
		wantCapped  int
	}{
		{
			name:        "上限に余裕があれば全量を付与する",
			amount:      10,
			usages:      []experienceCapUsage{{maxExperience: 100, usedExperience: 20}},
			wantAwarded: 10,
			wantCapped:  0,
		},
		{
			name:        "日次上限までの残量だけを付与する",
			amount:      10,
			usages:      []experienceCapUsage{{maxExperience: 100, usedExperience: 95}},
			wantAwarded: 5,
			wantCapped:  5,
		},
		{
			name:   "複数上限のうち最も厳しい残量を使う",
			amount: 10,
			usages: []experienceCapUsage{
				{maxExperience: 100, usedExperience: 92},
				{maxExperience: 500, usedExperience: 497},
			},
			wantAwarded: 3,
			wantCapped:  7,
		},
		{
			name:        "上限到達後は全量を制限する",
			amount:      10,
			usages:      []experienceCapUsage{{maxExperience: 100, usedExperience: 100}},
			wantAwarded: 0,
			wantCapped:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awarded, capped := calculateAwardedExperience(tt.amount, tt.usages)
			if awarded != tt.wantAwarded || capped != tt.wantCapped {
				t.Fatalf("awarded=%d, capped=%d; want awarded=%d, capped=%d", awarded, capped, tt.wantAwarded, tt.wantCapped)
			}
		})
	}
}

func TestExperiencePeriodStart(t *testing.T) {
	date := time.Date(2026, time.August, 27, 0, 0, 0, 0, timeutil.LocationJST())

	tests := []struct {
		name    string
		capType domain.ExperienceCapType
		want    string
	}{
		{name: "日次", capType: domain.ExperienceCapTypeDaily, want: "2026-08-27"},
		{name: "週次は月曜始まり", capType: domain.ExperienceCapTypeWeekly, want: "2026-08-24"},
		{name: "月次", capType: domain.ExperienceCapTypeMonthly, want: "2026-08-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := experiencePeriodStart(date, tt.capType).Format("2006-01-02")
			if got != tt.want {
				t.Fatalf("period start=%s; want %s", got, tt.want)
			}
		})
	}
}
