package domain

import "time"

type PetDepartureRuleID int

type PetDepartureRule struct {
	ID              PetDepartureRuleID
	RuleKey         string
	MinAgeDays      int
	RequiredStageID int
	RequiredStageNo int
	GraceDaysMin    int
	GraceDaysMax    int
}

type PetDepartureCandidate struct {
	PetID                PetID
	UserID               UserID
	CreatedAt            time.Time
	CurrentStageID       int
	CurrentStageNo       int
	StageReachedAt       *time.Time
	DepartureID          *string
	DepartureStatus      *string
	EligibleAt           *time.Time
	ScheduledDepartureAt *time.Time
}

type PetDeparture struct {
	Status               string
	EligibleAt           *time.Time
	ScheduledDepartureAt *time.Time
}

type PetDepartureUpsertInput struct {
	PetID                PetID
	UserID               UserID
	RuleID               PetDepartureRuleID
	Status               string
	EligibleAt           *time.Time
	ScheduledDepartureAt *time.Time
	BlockedReason        *string
	CheckedAt            time.Time
}

type PetDepartureDepartInput struct {
	PetID       PetID
	UserID      UserID
	RuleID      PetDepartureRuleID
	EligibleAt  time.Time
	ScheduledAt time.Time
	DepartedAt  time.Time
}

type PetDepartureRepository interface {
	FindActiveRule() (PetDepartureRule, error)
	FindActivePetsByUserID(rule PetDepartureRule, userID UserID) ([]PetDepartureCandidate, error)
	FindByPetID(petID PetID) (PetDeparture, error)
	Upsert(input PetDepartureUpsertInput) error
	Depart(input PetDepartureDepartInput) error
}
