package main

import (
	"testing"
	"time"
)

func TestRandomDayOfMonthIsValidAndStable(t *testing.T) {
	location := time.FixedZone("test", 9*60*60)

	for month := time.January; month <= time.December; month++ {
		got := randomDayOfMonth(2028, month, location)
		lastDay := time.Date(2028, month+1, 0, 0, 0, 0, 0, location).Day()
		if got < 1 || got > lastDay {
			t.Fatalf("month %d: day %d is outside 1..%d", month, got, lastDay)
		}
		if again := randomDayOfMonth(2028, month, location); again != got {
			t.Fatalf("month %d: day changed from %d to %d", month, got, again)
		}
	}
}

func TestNextMonthlyRandomDayRun(t *testing.T) {
	location := time.FixedZone("test", 9*60*60)
	year, month := 2028, time.February
	day := randomDayOfMonth(year, month, location)
	scheduled := time.Date(year, month, day, defaultMessageHour, 0, 0, 0, location)

	before := scheduled.Add(-time.Second)
	if got := nextMonthlyRandomDayRun(before, defaultMessageHour, location); !got.Equal(scheduled) {
		t.Fatalf("before scheduled time: got %s, want %s", got, scheduled)
	}

	after := scheduled.Add(time.Second)
	got := nextMonthlyRandomDayRun(after, defaultMessageHour, location)
	if got.Month() != time.March || got.Year() != year || got.Hour() != defaultMessageHour {
		t.Fatalf("after scheduled time: got %s, want a March run at %02d:00", got, defaultMessageHour)
	}
	wantDay := randomDayOfMonth(year, time.March, location)
	if got.Day() != wantDay {
		t.Fatalf("after scheduled time: got day %d, want %d", got.Day(), wantDay)
	}
}
