package persistence

import (
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/timeutil"
)

func TestPraiseReportDateValueUsesJSTCalendarDate(t *testing.T) {
	// UTC では前日でも、レポートの日付は JST の暦日を使う。
	reportDate := time.Date(2026, 8, 28, 15, 30, 0, 0, time.UTC)
	if got := praiseReportDateValue(reportDate); got != "2026-08-29" {
		t.Fatalf("praiseReportDateValue() = %q, want 2026-08-29", got)
	}

	if reportDate.In(timeutil.LocationJST()).Day() != 29 {
		t.Fatal("test fixture must cross the JST calendar boundary")
	}
}
