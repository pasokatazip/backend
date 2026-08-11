package usecases

import (
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type subscriptionReportRepoStub struct {
	reports []domain.Report
	userID  domain.UserID
	date    time.Time
}

func (s *subscriptionReportRepoStub) FindByUserAndDate(userID domain.UserID, date time.Time) ([]domain.Report, error) {
	s.userID, s.date = userID, date
	return s.reports, nil
}

type subscriptionReportPetRepoStub struct {
	pet   domain.Pet
	petID domain.PetID
}

func (s *subscriptionReportPetRepoStub) FindByID(petID domain.PetID) (domain.Pet, error) {
	s.petID = petID
	return s.pet, nil
}

func TestFindSubscriptionReportsDerivesPetIDFromReport(t *testing.T) {
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	petID := domain.PetID("b5d213dd-75f7-4bb2-b260-7efb4c04758a")
	date := time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)
	gossip := "散歩した"
	report, err := domain.NewPersistedReport(
		domain.ReportID("5ec57584-b8bd-4b94-b060-43bf7d516b1e"), petID, 10, &gossip,
		1, "stayed", "のんびり", "公園", date, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	pet := domain.NewPet(petID, "ぽち", "#FFC1CA", false, userID, 1, 2, 3, 4, nil, 2, date, date)
	reportRepo := &subscriptionReportRepoStub{reports: []domain.Report{report}}
	petRepo := &subscriptionReportPetRepoStub{pet: pet}

	got, err := NewFindSubscriptionReports(reportRepo, petRepo).Execute(FindSubscriptionReportsInput{
		UserID: userID,
		Date:   date,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if reportRepo.userID != userID || !reportRepo.date.Equal(date) {
		t.Fatalf("report query = (%s, %s)", reportRepo.userID, reportRepo.date)
	}
	if petRepo.petID != petID {
		t.Fatalf("pet query id = %s, want %s", petRepo.petID, petID)
	}
	if len(got.Reports) != 1 || got.Pet.ID != string(petID) || got.Pet.Name != "ぽち" {
		t.Fatalf("output = %+v", got)
	}
}

func TestFindSubscriptionReportsReturnsNotFoundWithEmptyReports(t *testing.T) {
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	now := time.Now()
	reportRepo := &subscriptionReportRepoStub{}
	petRepo := &subscriptionReportPetRepoStub{}

	_, err := NewFindSubscriptionReports(reportRepo, petRepo).Execute(FindSubscriptionReportsInput{
		UserID: userID,
		Date:   now,
	})
	if err != domain.ErrNotFound {
		t.Fatalf("Execute() error = %v, want %v", err, domain.ErrNotFound)
	}
	if petRepo.petID != "" {
		t.Fatalf("pet repository called with %s", petRepo.petID)
	}
}
