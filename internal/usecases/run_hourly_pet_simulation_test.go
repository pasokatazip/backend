package usecases

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type dailyPropagationLimitRepositoryStub struct {
	candidates     []domain.InterestPropagationCandidate
	saved          []domain.PetInterestPropagation
	reportGroupIDs []domain.GroupMasterID
	recentVisits   domain.PetGroupVisitCounts
	maxDailySaves  int
	reportsCreated int
}

func (r *dailyPropagationLimitRepositoryStub) FindActivePetsForSimulation() ([]domain.SimulationPet, error) {
	return nil, nil
}

func (r *dailyPropagationLimitRepositoryStub) FindActiveGroupsForSimulation() ([]domain.GroupMaster, error) {
	return []domain.GroupMaster{
		domain.NewGroupMaster(1, "park", "公園の群れ", nil, 0, 0, 0, 0, 0, 1, 1, 1, true, time.Now()),
	}, nil
}

func (r *dailyPropagationLimitRepositoryStub) PruneExpiredGroupInterestsForSimulation() error {
	return nil
}

func (r *dailyPropagationLimitRepositoryStub) FindGroupInterestsForSimulation() (domain.PetGroupInterests, error) {
	return domain.PetGroupInterests{}, nil
}

func (r *dailyPropagationLimitRepositoryStub) FindRecentGroupVisitCountsForSimulation(
	_ time.Time,
) (domain.PetGroupVisitCounts, error) {
	return r.recentVisits, nil
}

func (r *dailyPropagationLimitRepositoryStub) FindInterestPropagationCandidates(
	_ time.Time,
) ([]domain.InterestPropagationCandidate, error) {
	return r.candidates, nil
}

func (r *dailyPropagationLimitRepositoryStub) SaveInterestPropagation(
	propagation domain.PetInterestPropagation,
) (bool, error) {
	// DBトリガーが上限到達時に INSERT をスキップする挙動を再現する。
	if len(r.saved) >= r.maxDailySaves {
		return false, nil
	}
	r.saved = append(r.saved, propagation)
	return true, nil
}

func (r *dailyPropagationLimitRepositoryStub) AppendInterestPropagationReportMaterial(
	_ domain.PetID,
	_ time.Time,
	propagatedGroupID domain.GroupMasterID,
) error {
	r.reportGroupIDs = append(r.reportGroupIDs, propagatedGroupID)
	return nil
}

func (r *dailyPropagationLimitRepositoryStub) SaveHourlySimulation(
	_ domain.PetSimulationSaveInput,
) (bool, error) {
	return true, nil
}

func (r *dailyPropagationLimitRepositoryStub) CreateReportsForSimulation(_ time.Time) (int, error) {
	return r.reportsCreated, nil
}

func TestRunHourlyPetSimulationDoesNotReportPropagationRejectedByDailyLimit(t *testing.T) {
	recipientPetID := domain.PetID("recipient-pet")
	repo := &dailyPropagationLimitRepositoryStub{
		maxDailySaves: 2,
		candidates: []domain.InterestPropagationCandidate{
			{
				RecipientPetID: recipientPetID, SourcePetID: "source-pet-1", SourceHourlyLogID: "log-1",
				ViaGroupMasterID: 1, PropagatedGroupMasterID: 11, SourceInterestScore: 0.5,
				SourceSociality: 50, RecipientCuriosity: 50,
			},
			{
				RecipientPetID: recipientPetID, SourcePetID: "source-pet-2", SourceHourlyLogID: "log-2",
				ViaGroupMasterID: 1, PropagatedGroupMasterID: 12, SourceInterestScore: 5,
				SourceSociality: 50, RecipientCuriosity: 50,
			},
			{
				// 最も強い候補でも、1日の上限で保存されなければ表示しない。
				RecipientPetID: recipientPetID, SourcePetID: "source-pet-3", SourceHourlyLogID: "log-3",
				ViaGroupMasterID: 1, PropagatedGroupMasterID: 13, SourceInterestScore: 100,
				SourceSociality: 100, RecipientCuriosity: 100,
			},
		},
	}
	simulatedAt := time.Date(2026, 9, 3, 12, 0, 0, 0, time.FixedZone("JST", 9*60*60))

	output, err := NewRunHourlyPetSimulation(repo).Execute(RunHourlyPetSimulationInput{
		SimulatedAt: &simulatedAt,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.InterestPropagations != 2 {
		t.Fatalf("InterestPropagations = %d, want 2", output.InterestPropagations)
	}
	if len(repo.saved) != 2 {
		t.Fatalf("saved propagations = %d, want 2", len(repo.saved))
	}
	if len(repo.reportGroupIDs) != 1 || repo.reportGroupIDs[0] != 12 {
		t.Fatalf("report group IDs = %v, want [12]", repo.reportGroupIDs)
	}
}

func TestRunHourlyPetSimulationDoesNotAcquireCurrentGroupPropagation(t *testing.T) {
	recipientPetID := domain.PetID("recipient-pet")
	repo := &dailyPropagationLimitRepositoryStub{
		maxDailySaves: 2,
		candidates: []domain.InterestPropagationCandidate{
			{
				RecipientPetID: recipientPetID, SourcePetID: "source-pet", SourceHourlyLogID: "log",
				ViaGroupMasterID: 1, PropagatedGroupMasterID: 1, SourceInterestScore: 100,
				SourceSociality: 100, RecipientCuriosity: 100,
			},
		},
	}
	simulatedAt := time.Date(2026, 9, 3, 13, 0, 0, 0, time.FixedZone("JST", 9*60*60))

	output, err := NewRunHourlyPetSimulation(repo).Execute(RunHourlyPetSimulationInput{
		SimulatedAt: &simulatedAt,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.InterestPropagations != 0 {
		t.Fatalf("InterestPropagations = %d, want 0", output.InterestPropagations)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("saved propagations = %d, want 0", len(repo.saved))
	}
	if len(repo.reportGroupIDs) != 0 {
		t.Fatalf("report group IDs = %v, want none", repo.reportGroupIDs)
	}
}

func TestRecentGroupVisitPenaltyGrowsAndHasUpperBound(t *testing.T) {
	if got := recentGroupVisitPenalty(0); got != 0 {
		t.Fatalf("recentGroupVisitPenalty(0) = %f, want 0", got)
	}

	oneVisit := recentGroupVisitPenalty(1)
	manyVisits := recentGroupVisitPenalty(100)
	if oneVisit <= 0 || manyVisits <= oneVisit {
		t.Fatalf("penalties = (%f, %f), want increasing positive values", oneVisit, manyVisits)
	}
	if manyVisits > maxRecentGroupVisitPenalty || math.Abs(manyVisits-maxRecentGroupVisitPenalty) > 0.0001 {
		t.Fatalf("recentGroupVisitPenalty(100) = %f, want approximately %f", manyVisits, maxRecentGroupVisitPenalty)
	}
}

func TestCandidateSelectionWeightReducesFrequentlyVisitedGroup(t *testing.T) {
	fresh := nextGroupCandidate{timeWeight: 1}
	repeated := nextGroupCandidate{
		timeWeight:   1,
		visitPenalty: maxRecentGroupVisitPenalty,
	}

	freshWeight := candidateSelectionWeight(fresh)
	repeatedWeight := candidateSelectionWeight(repeated)
	if repeatedWeight >= freshWeight {
		t.Fatalf("weights = repeated %f, fresh %f; repeated group must be less likely", repeatedWeight, freshWeight)
	}
	if math.Abs(repeatedWeight-recentGroupVisitWeightFloor) > 0.0001 {
		t.Fatalf("repeated weight = %f, want floor %f", repeatedWeight, recentGroupVisitWeightFloor)
	}
}

func TestCloseCandidatePoolSizeIncludesMoreComparableGroups(t *testing.T) {
	candidates := []nextGroupCandidate{
		{score: 1.00},
		{score: 0.96},
		{score: 0.92},
		{score: 0.88},
		{score: 0.84},
		{score: 0.83},
		{score: 0.81},
	}

	if got := closeCandidatePoolSize(candidates); got != maxCloseGroupCandidatePool {
		t.Fatalf("closeCandidatePoolSize() = %d, want %d", got, maxCloseGroupCandidatePool)
	}
}

func TestChooseNextGroupAvoidsHeavilyVisitedGroup(t *testing.T) {
	category := "hobby"
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	currentGroupID := 99
	pet := domain.SimulationPet{Pet: domain.NewPet(
		"pet", "ペット", domain.DefaultPetColor, false, "user",
		50, 50, 50, 50, &currentGroupID, 1, now, now,
	)}
	groups := []domain.GroupMaster{
		domain.NewGroupMaster(1, "repeated", "何度も行った群れ", &category, 0, 0, 0, 0, 0, 1, 1, 1, true, now),
		domain.NewGroupMaster(2, "fresh", "まだ行っていない群れ", &category, 0, 0, 0, 0, 0, 1, 1, 1, true, now),
	}

	for seed := int64(0); seed < 20; seed++ {
		selected := chooseNextGroup(
			pet,
			groups,
			99,
			0,
			nil,
			domain.GroupVisitCounts{1: 12},
			now,
			rand.New(rand.NewSource(seed)),
		)
		if selected.ID() != 2 {
			t.Fatalf("seed %d selected group %d, want fresh group 2", seed, selected.ID())
		}
	}
}
