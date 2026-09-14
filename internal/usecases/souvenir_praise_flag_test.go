package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type markPraiseRepositoryStub struct {
	marked     bool
	reportDate time.Time
}

func (r *markPraiseRepositoryStub) FindByPetIDAndDate(_ context.Context, _ domain.PetID, _ time.Time) (domain.SouvenirPraiseFlag, error) {
	return domain.SouvenirPraiseFlag{}, nil
}

func (r *markPraiseRepositoryStub) MarkPraised(_ context.Context,
	userID domain.UserID,
	reportDate time.Time,
) (domain.SouvenirPraiseFlag, error) {
	r.marked = true
	r.reportDate = reportDate
	praisedAt := timeutil.NowJST()
	return domain.NewSouvenirPraiseFlag(userID, reportDate, true, &praisedAt), nil
}

type markPraiseReportRepositoryStub struct {
	domain.ReportRepository
	called     bool
	userID     domain.UserID
	reportDate time.Time
	reports    []domain.Report
	err        error
}

func (r *markPraiseReportRepositoryStub) FindByUserAndDate(_ context.Context,
	userID domain.UserID,
	reportDate time.Time,
) ([]domain.Report, error) {
	r.called = true
	r.userID = userID
	r.reportDate = reportDate
	return r.reports, r.err
}

func newPraiseTestReport(t *testing.T, withSouvenir bool) domain.Report {
	t.Helper()
	gossip := "gossip"
	report, err := domain.NewPersistedReport(
		domain.ReportID("report-id"),
		domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210"),
		12,
		&gossip,
		1,
		"behavior",
		"label",
		"group",
		timeutil.NowJST(),
		nil,
	)
	if err != nil {
		t.Fatalf("NewPersistedReport: %v", err)
	}
	if withSouvenir {
		report = report.WithSouvenirs([]domain.ReportSouvenir{
			domain.NewReportSouvenir("souvenir-id", "souvenir", "https://example.com/image.png"),
		})
	}
	return report
}

func TestMarkSouvenirPraisedNormalizesJSTDate(t *testing.T) {
	userID := domain.UserID("c9428888-122b-11e1-b85c-61cd3cbb3210")
	inputDate := timeutil.NowJST().Add(-12 * time.Hour)
	expectedDate := normalizeJSTDate(inputDate)
	reportRepo := &markPraiseReportRepositoryStub{reports: []domain.Report{newPraiseTestReport(t, true)}}
	praiseRepo := &markPraiseRepositoryStub{}

	output, err := NewMarkSouvenirPraised(praiseRepo, reportRepo).Execute(context.Background(), MarkSouvenirPraisedInput{
		UserID:     userID,
		ReportDate: inputDate,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !reportRepo.reportDate.Equal(expectedDate) || !praiseRepo.reportDate.Equal(expectedDate) {
		t.Fatalf("dates = (%s, %s), want %s", reportRepo.reportDate, praiseRepo.reportDate, expectedDate)
	}
	if !output.ReportDate.Equal(expectedDate) || !output.HasPraised {
		t.Fatalf("output = %+v, want praised on %s", output, expectedDate)
	}
}

func TestMarkSouvenirPraisedRejectsFutureJSTDate(t *testing.T) {
	reportRepo := &markPraiseReportRepositoryStub{}
	praiseRepo := &markPraiseRepositoryStub{}

	_, err := NewMarkSouvenirPraised(praiseRepo, reportRepo).Execute(context.Background(), MarkSouvenirPraisedInput{
		UserID:     "c9428888-122b-11e1-b85c-61cd3cbb3210",
		ReportDate: normalizeJSTDate(timeutil.NowJST()).AddDate(0, 0, 1),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
	if reportRepo.called || praiseRepo.marked {
		t.Fatal("repositories were called for a future report date")
	}
}

func TestMarkSouvenirPraisedRequiresReportWithSouvenir(t *testing.T) {
	tests := []struct {
		name    string
		reports []domain.Report
	}{
		{name: "report does not exist"},
		{name: "report has no souvenir", reports: []domain.Report{newPraiseTestReport(t, false)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportRepo := &markPraiseReportRepositoryStub{reports: tt.reports}
			praiseRepo := &markPraiseRepositoryStub{}

			_, err := NewMarkSouvenirPraised(praiseRepo, reportRepo).Execute(context.Background(), MarkSouvenirPraisedInput{
				UserID:     "c9428888-122b-11e1-b85c-61cd3cbb3210",
				ReportDate: timeutil.NowJST(),
			})
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("error = %v, want ErrNotFound", err)
			}
			if praiseRepo.marked {
				t.Fatal("praise flag was written without a report containing a souvenir")
			}
		})
	}
}
