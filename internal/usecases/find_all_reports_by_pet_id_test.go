package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type findAllReportsRepository struct {
	petID       domain.PetID
	reports     []domain.Report
	dateReports []domain.Report
	reportDate  time.Time
}

type findByDatePraiseRepository struct {
	flag       domain.SouvenirPraiseFlag
	petID      domain.PetID
	reportDate time.Time
}

type reportPetRepositoryStub struct {
	domain.PetRepository
	pet domain.Pet
	err error
}

func (r *reportPetRepositoryStub) FindByID(domain.PetID) (domain.Pet, error) {
	return r.pet, r.err
}

func newOwnedReportPet(petID domain.PetID, userID domain.UserID) domain.Pet {
	return domain.NewPet(
		petID, "pet", domain.DefaultPetColor, false, userID,
		0, 0, 0, 0, nil, 1, time.Time{}, time.Time{},
	)
}

func (r *findByDatePraiseRepository) FindByPetIDAndDate(
	petID domain.PetID,
	reportDate time.Time,
) (domain.SouvenirPraiseFlag, error) {
	r.petID = petID
	r.reportDate = reportDate
	return r.flag, nil
}

func (r *findByDatePraiseRepository) MarkPraised(
	_ domain.UserID,
	_ time.Time,
) (domain.SouvenirPraiseFlag, error) {
	return r.flag, nil
}

func (r *findAllReportsRepository) FindByDate(_ domain.PetID, reportDate time.Time) ([]domain.Report, error) {
	r.reportDate = reportDate
	return r.dateReports, nil
}

func newReportForOutputTest(t *testing.T, petID domain.PetID, groupName string) domain.Report {
	t.Helper()
	gossip := "gossip"
	report, err := domain.NewPersistedReport(
		domain.ReportID("report-id"), petID, 12, &gossip, 42,
		"behavior", "label", groupName, time.Now(),
		[]string{"近くでゲームの話をしていた"},
	)
	if err != nil {
		t.Fatalf("NewPersistedReport: %v", err)
	}
	return report
}

func (r *findAllReportsRepository) FindAllByPetID(petID domain.PetID) ([]domain.Report, error) {
	r.petID = petID
	return r.reports, nil
}

func TestFindAllReportsByPetID(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	userID := domain.UserID("c9428888-122b-11e1-b85c-61cd3cbb3210")
	repo := &findAllReportsRepository{reports: []domain.Report{newReportForOutputTest(t, petID, "公園の群れ")}}
	petRepo := &reportPetRepositoryStub{pet: newOwnedReportPet(petID, userID)}

	outputs, err := NewFindAllReportsByPetID(repo, petRepo).Execute(FindAllReportsByPetIDInput{
		UserID: userID,
		PetID:  petID,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.petID != petID {
		t.Fatalf("repository pet ID = %q, want %q", repo.petID, petID)
	}
	if len(outputs) != 1 || outputs[0].GroupName != "公園の群れ" {
		t.Fatalf("outputs = %+v, want GroupName 公園の群れ", outputs)
	}
}

func TestFindByDateReportIncludesGroupMasterID(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	userID := domain.UserID("c9428888-122b-11e1-b85c-61cd3cbb3210")
	report := newReportForOutputTest(t, petID, "駅前の群れ")
	report = report.WithSouvenirs([]domain.ReportSouvenir{
		domain.NewReportSouvenir("souvenir-id", "おみやげ", "https://example.com/souvenir.png"),
	})
	repo := &findAllReportsRepository{
		dateReports: []domain.Report{report},
	}
	reportDate := time.Date(2026, time.July, 20, 0, 0, 0, 0, timeutil.LocationJST())
	praiseRepo := &findByDatePraiseRepository{
		flag: domain.NewSouvenirPraiseFlag(
			userID,
			reportDate,
			true,
			&reportDate,
		),
	}

	petRepo := &reportPetRepositoryStub{pet: newOwnedReportPet(petID, userID)}
	output, err := NewFindByDate(repo, praiseRepo, petRepo).Execute(FindByDateReportInput{
		UserID:     userID,
		PetID:      petID,
		ReportDate: &reportDate,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(output.Reports) != 1 || output.Reports[0].GroupName != "駅前の群れ" {
		t.Fatalf("reports = %+v, want GroupName 駅前の群れ", output.Reports)
	}
	if len(output.Reports[0].Souvenirs) != 1 || output.Reports[0].Souvenirs[0].ID != "souvenir-id" {
		t.Fatalf("souvenirs = %+v, want souvenir-id", output.Reports[0].Souvenirs)
	}
	if len(output.Reports[0].Rumors) != 1 || output.Reports[0].Rumors[0] != "近くでゲームの話をしていた" {
		t.Fatalf("rumors = %+v, want one rumor", output.Reports[0].Rumors)
	}
	if !output.HasPraised {
		t.Fatal("HasPraised = false, want true")
	}
	if !repo.reportDate.Equal(reportDate) {
		t.Fatalf("report date = %s, want %s", repo.reportDate, reportDate)
	}
	if praiseRepo.petID != petID || !praiseRepo.reportDate.Equal(reportDate) {
		t.Fatalf("praise lookup = (%s, %s), want (%s, %s)", praiseRepo.petID, praiseRepo.reportDate, petID, reportDate)
	}
}

func TestDefaultReportDateIsPreviousJSTDay(t *testing.T) {
	now := time.Date(2026, time.July, 23, 10, 30, 0, 0, timeutil.LocationJST())
	want := time.Date(2026, time.July, 22, 10, 30, 0, 0, timeutil.LocationJST())

	if got := defaultReportDate(now); !got.Equal(want) {
		t.Fatalf("default report date = %s, want %s", got, want)
	}
}

func TestFindAllReportsByPetIDRejectsInvalidPetID(t *testing.T) {
	_, err := NewFindAllReportsByPetID(
		&findAllReportsRepository{},
		&reportPetRepositoryStub{},
	).Execute(
		FindAllReportsByPetIDInput{
			UserID: "c9428888-122b-11e1-b85c-61cd3cbb3210",
			PetID:  "invalid",
		},
	)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
}

func TestFindAllReportsByPetIDRejectsAnotherUsersPet(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	requestUserID := domain.UserID("c9428888-122b-11e1-b85c-61cd3cbb3210")
	ownerID := domain.UserID("a9428888-122b-11e1-b85c-61cd3cbb3210")
	reportRepo := &findAllReportsRepository{}
	petRepo := &reportPetRepositoryStub{pet: newOwnedReportPet(petID, ownerID)}

	_, err := NewFindAllReportsByPetID(reportRepo, petRepo).Execute(
		FindAllReportsByPetIDInput{UserID: requestUserID, PetID: petID},
	)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("error = %v, want ErrUnauthorized", err)
	}
	if reportRepo.petID != "" {
		t.Fatal("report repository was called before ownership validation")
	}
}

func TestFindByDateReportRejectsAnotherUsersPet(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	requestUserID := domain.UserID("c9428888-122b-11e1-b85c-61cd3cbb3210")
	ownerID := domain.UserID("a9428888-122b-11e1-b85c-61cd3cbb3210")
	reportRepo := &findAllReportsRepository{}
	petRepo := &reportPetRepositoryStub{pet: newOwnedReportPet(petID, ownerID)}

	_, err := NewFindByDate(reportRepo, &findByDatePraiseRepository{}, petRepo).Execute(
		FindByDateReportInput{UserID: requestUserID, PetID: petID},
	)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("error = %v, want ErrUnauthorized", err)
	}
	if !reportRepo.reportDate.IsZero() {
		t.Fatal("report repository was called before ownership validation")
	}
}
