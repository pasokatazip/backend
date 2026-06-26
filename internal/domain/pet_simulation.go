package domain

import "time"

type SimulationPet struct {
	Pet
	CurrentJoinID *string
	JoinedAt      *time.Time
}

type PetSimulationSaveInput struct {
	PetID            PetID
	PreviousGroupID  *GroupMasterID
	NextGroupID      GroupMasterID
	PreviousJoinID   *string
	MoveReason       string
	Moved            bool
	EnergyDelta      float64
	CuriosityDelta   float64
	SocialityDelta   float64
	RoutineDelta     float64
	ExperienceAmount int
	SimulatedAt      time.Time
	Log              PetHourlyLog
}

type PetSimulationRepository interface {
	FindActivePetsForSimulation() ([]SimulationPet, error)
	FindActiveGroupsForSimulation() ([]GroupMaster, error)
	SaveHourlySimulation(input PetSimulationSaveInput) (bool, error)
}
