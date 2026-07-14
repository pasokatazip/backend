package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type findAllReportsRepository struct {
	petID        domain.PetID
	reports      []domain.Report
	todayReports []domain.Report
}

func (r *findAllReportsRepository) FindByToday(domain.PetID) ([]domain.Report, error) {
	return r.todayReports, nil
}

func newReportForOutputTest(t *testing.T, petID domain.PetID, groupName string) domain.Report {
	t.Helper()
	gossip := "gossip"
	report, err := domain.NewPersistedReport(
		domain.ReportID("report-id"), petID, 12, &gossip, 42,
		"behavior", "label", groupName, time.Now(),
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
	repo := &findAllReportsRepository{reports: []domain.Report{newReportForOutputTest(t, petID, "公園の群れ")}}

	outputs, err := NewFindAllReportsByPetID(repo).Execute(FindAllReportsByPetIDInput{PetID: petID})
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

func TestFindByTodayReportIncludesGroupMasterID(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	report := newReportForOutputTest(t, petID, "駅前の群れ")
	report = report.WithSouvenirs([]domain.ReportSouvenir{
		domain.NewReportSouvenir("souvenir-id", "おみやげ", "https://example.com/souvenir.png"),
	})
	repo := &findAllReportsRepository{
		todayReports: []domain.Report{report},
	}

	outputs, err := NewFindByToDay(repo).Execute(FindByTodayReportInput{PetID: petID})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(outputs) != 1 || outputs[0].GroupName != "駅前の群れ" {
		t.Fatalf("outputs = %+v, want GroupName 駅前の群れ", outputs)
	}
	if len(outputs[0].Souvenirs) != 1 || outputs[0].Souvenirs[0].ID != "souvenir-id" {
		t.Fatalf("souvenirs = %+v, want souvenir-id", outputs[0].Souvenirs)
	}
}

func TestFindAllReportsByPetIDRejectsInvalidPetID(t *testing.T) {
	_, err := NewFindAllReportsByPetID(&findAllReportsRepository{}).Execute(
		FindAllReportsByPetIDInput{PetID: "invalid"},
	)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
}
