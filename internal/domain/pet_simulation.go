package domain

import "time"

type GroupInterestScores map[GroupMasterID]float64

type PetGroupInterests map[PetID]GroupInterestScores

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
	FindGroupInterestsForSimulation() (PetGroupInterests, error)
	SaveHourlySimulation(input PetSimulationSaveInput) (bool, error)
}
