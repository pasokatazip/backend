package usecases

import (
	"errors"
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

type subscriptionPraiseRepoStub struct {
	flag       domain.SouvenirPraiseFlag
	petID      domain.PetID
	reportDate time.Time
}

type subscriptionEvolutionStageRepoStub struct {
	stage domain.EvolutionStage
	id    domain.EvolutionStageID
	err   error
}

func (s *subscriptionEvolutionStageRepoStub) FindByID(
	id domain.EvolutionStageID,
) (domain.EvolutionStage, error) {
	s.id = id
	return s.stage, s.err
}

func (s *subscriptionEvolutionStageRepoStub) FindByStageNo(int) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (s *subscriptionEvolutionStageRepoStub) FindAll() ([]domain.EvolutionStage, error) {
	return nil, nil
}

func (s *subscriptionPraiseRepoStub) FindByPetIDAndDate(
	petID domain.PetID,
	reportDate time.Time,
) (domain.SouvenirPraiseFlag, error) {
	s.petID = petID
	s.reportDate = reportDate
	return s.flag, nil
}

func (s *subscriptionPraiseRepoStub) MarkPraised(
	_ domain.UserID,
	_ time.Time,
) (domain.SouvenirPraiseFlag, error) {
	return s.flag, nil
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
	stage := domain.NewEvolutionStage(2, "namai_shokushu", 2, "なまい期・しょくしゅ", nil, nil, date, date)
	reportRepo := &subscriptionReportRepoStub{reports: []domain.Report{report}}
	petRepo := &subscriptionReportPetRepoStub{pet: pet}
	stageRepo := &subscriptionEvolutionStageRepoStub{stage: stage}
	praiseRepo := &subscriptionPraiseRepoStub{
		flag: domain.NewSouvenirPraiseFlag(userID, date, true, &date),
	}

	got, err := NewFindSubscriptionReports(reportRepo, petRepo, stageRepo, praiseRepo).Execute(FindSubscriptionReportsInput{
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
	if stageRepo.id != 2 {
		t.Fatalf("stage query id = %d, want 2", stageRepo.id)
	}
	if praiseRepo.petID != petID || !praiseRepo.reportDate.Equal(date) {
		t.Fatalf("praise query = (%s, %s), want (%s, %s)", praiseRepo.petID, praiseRepo.reportDate, petID, date)
	}
	if len(got.Reports) != 1 || got.Pet.ID != string(petID) || got.Pet.Name != "ぽち" {
		t.Fatalf("output = %+v", got)
	}
	if got.Pet.CurrentStageKey != "namai_shokushu" || got.Pet.CurrentStageNo != 2 {
		t.Fatalf("pet evolution stage = (%s, %d), want (namai_shokushu, 2)", got.Pet.CurrentStageKey, got.Pet.CurrentStageNo)
	}
	if !got.HasPraised {
		t.Fatal("HasPraised = false, want true")
	}
}

func TestFindSubscriptionReportsReturnsNotFoundWithEmptyReports(t *testing.T) {
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	now := time.Now()
	reportRepo := &subscriptionReportRepoStub{}
	petRepo := &subscriptionReportPetRepoStub{}
	stageRepo := &subscriptionEvolutionStageRepoStub{}
	praiseRepo := &subscriptionPraiseRepoStub{}

	_, err := NewFindSubscriptionReports(reportRepo, petRepo, stageRepo, praiseRepo).Execute(FindSubscriptionReportsInput{
		UserID: userID,
		Date:   now,
	})
	if err != domain.ErrNotFound {
		t.Fatalf("Execute() error = %v, want %v", err, domain.ErrNotFound)
	}
	if petRepo.petID != "" {
		t.Fatalf("pet repository called with %s", petRepo.petID)
	}
	if praiseRepo.petID != "" {
		t.Fatalf("praise repository called with %s", praiseRepo.petID)
	}
	if stageRepo.id != 0 {
		t.Fatalf("stage repository called with %d", stageRepo.id)
	}
}

func TestFindSubscriptionReportsReturnsEvolutionStageLookupError(t *testing.T) {
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
	stageErr := errors.New("stage lookup failed")
	reportRepo := &subscriptionReportRepoStub{reports: []domain.Report{report}}
	petRepo := &subscriptionReportPetRepoStub{
		pet: domain.NewPet(petID, "ぽち", "#FFC1CA", false, userID, 1, 2, 3, 4, nil, 2, date, date),
	}
	stageRepo := &subscriptionEvolutionStageRepoStub{err: stageErr}
	praiseRepo := &subscriptionPraiseRepoStub{}

	_, err = NewFindSubscriptionReports(reportRepo, petRepo, stageRepo, praiseRepo).Execute(
		FindSubscriptionReportsInput{UserID: userID, Date: date},
	)
	if !errors.Is(err, stageErr) {
		t.Fatalf("Execute() error = %v, want %v", err, stageErr)
	}
	if praiseRepo.petID != "" {
		t.Fatalf("praise repository called with %s after stage lookup failed", praiseRepo.petID)
	}
}
