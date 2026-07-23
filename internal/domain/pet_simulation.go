package domain

import "time"

type GroupInterestScores map[GroupMasterID]float64

type PetGroupInterests map[PetID]GroupInterestScores

// 同じ時刻・同じ群れにいたペット同士で興味を伝えられる組み合わせ 投稿本文や名詞そのものは含めない
type InterestPropagationCandidate struct {
	RecipientPetID          PetID
	SourcePetID             PetID
	SourceHourlyLogID       PetHourlyLogID
	ViaGroupMasterID        GroupMasterID
	PropagatedGroupMasterID GroupMasterID
	SourceInterestScore     float64
	SourceSociality         float64
	RecipientCuriosity      float64
}

// 1回の興味伝播を保存するための値
// 保存済みかどうかは source hourly log と受け手・伝播先群れの組み合わせで判定
type PetInterestPropagation struct {
	RecipientPetID          PetID
	SourcePetID             PetID
	SourceHourlyLogID       PetHourlyLogID
	ViaGroupMasterID        GroupMasterID
	PropagatedGroupMasterID GroupMasterID
	Amount                  float64
	OccurredAt              time.Time
}

type SimulationPet struct {
	Pet
	CurrentJoinID *string
	JoinedAt      *time.Time
}

type PetSimulationSaveInput struct {
	PetID           PetID
	PreviousGroupID *GroupMasterID
	NextGroupID     GroupMasterID
	PreviousJoinID  *string
	MoveReason      string
	Moved           bool
	EnergyDelta     float64
	CuriosityDelta  float64
	SocialityDelta  float64
	RoutineDelta    float64
	SimulatedAt     time.Time
	Log             PetHourlyLog
	SouvenirDrop    bool
	SouvenirNote    string
}

type PetSimulationRepository interface {
	FindActivePetsForSimulation() ([]SimulationPet, error)
	FindActiveGroupsForSimulation() ([]GroupMaster, error)
	PruneExpiredGroupInterestsForSimulation() error
	FindGroupInterestsForSimulation() (PetGroupInterests, error)
	FindInterestPropagationCandidates(simulatedAt time.Time) ([]InterestPropagationCandidate, error)
	SaveInterestPropagation(propagation PetInterestPropagation) (bool, error)
	AppendInterestPropagationReportMaterial(petID PetID, simulatedAt time.Time, propagatedGroupID GroupMasterID) error
	SaveHourlySimulation(input PetSimulationSaveInput) (bool, error)
	CreateReportsForSimulation(simulatedAt time.Time) (int, error)
}
