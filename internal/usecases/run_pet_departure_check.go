package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

const (
	petDepartureStatusWaiting   = "waiting"
	petDepartureStatusScheduled = "scheduled"
	petDepartureStatusDeparted  = "departed"
	petDepartureStatusBlocked   = "blocked"

	petDepartureBlockedStageNotReached = "required_stage_not_reached"
)

type RunPetDepartureCheckInput struct {
	CheckedAt *time.Time
	UserID    domain.UserID
}

type RunPetDepartureCheckOutput struct {
	CheckedAt time.Time                       `json:"checked_at"`
	TotalPets int                             `json:"total_pets"`
	Waiting   int                             `json:"waiting"`
	Blocked   int                             `json:"blocked"`
	Scheduled int                             `json:"scheduled"`
	Departed  int                             `json:"departed"`
	Results   []RunPetDepartureCheckPetResult `json:"results"`
}

type RunPetDepartureCheckPetResult struct {
	PetID                string     `json:"pet_id"`
	Status               string     `json:"status"`
	EligibleAt           *time.Time `json:"eligible_at,omitempty"`
	ScheduledDepartureAt *time.Time `json:"scheduled_departure_at,omitempty"`
	DepartedAt           *time.Time `json:"departed_at,omitempty"`
	BlockedReason        *string    `json:"blocked_reason,omitempty"`
}

type RunPetDepartureCheck struct {
	repo domain.PetDepartureRepository
}

func NewRunPetDepartureCheck(repo domain.PetDepartureRepository) *RunPetDepartureCheck {
	return &RunPetDepartureCheck{repo: repo}
}

func (u *RunPetDepartureCheck) Execute(input RunPetDepartureCheckInput) (RunPetDepartureCheckOutput, error) {
	if !domain.IsValidUserID(input.UserID) {
		return RunPetDepartureCheckOutput{}, domain.ErrValidation
	}

	checkedAt := timeutil.NowJST()
	if input.CheckedAt != nil {
		checkedAt = input.CheckedAt.In(timeutil.LocationJST())
	}

	rule, err := u.repo.FindActiveRule()
	if err != nil {
		return RunPetDepartureCheckOutput{}, err
	}

	pets, err := u.repo.FindActivePetsByUserID(rule, input.UserID)
	if err != nil {
		return RunPetDepartureCheckOutput{}, err
	}

	output := RunPetDepartureCheckOutput{
		CheckedAt: checkedAt,
		TotalPets: len(pets),
		Results:   make([]RunPetDepartureCheckPetResult, 0, len(pets)),
	}

	for _, pet := range pets {
		result, err := u.checkPet(rule, pet, checkedAt)
		if err != nil {
			return RunPetDepartureCheckOutput{}, err
		}

		switch result.Status {
		case petDepartureStatusWaiting:
			output.Waiting++
		case petDepartureStatusBlocked:
			output.Blocked++
		case petDepartureStatusScheduled:
			output.Scheduled++
		case petDepartureStatusDeparted:
			output.Departed++
		}
		output.Results = append(output.Results, result)
	}

	return output, nil
}

func (u *RunPetDepartureCheck) checkPet(rule domain.PetDepartureRule, pet domain.PetDepartureCandidate, checkedAt time.Time) (RunPetDepartureCheckPetResult, error) {
	minAgeAt := pet.CreatedAt.In(timeutil.LocationJST()).AddDate(0, 0, rule.MinAgeDays)
	if checkedAt.Before(minAgeAt) {
		if err := u.repo.Upsert(domain.PetDepartureUpsertInput{
			PetID:     pet.PetID,
			UserID:    pet.UserID,
			RuleID:    rule.ID,
			Status:    petDepartureStatusWaiting,
			CheckedAt: checkedAt,
		}); err != nil {
			return RunPetDepartureCheckPetResult{}, err
		}
		return RunPetDepartureCheckPetResult{
			PetID:  string(pet.PetID),
			Status: petDepartureStatusWaiting,
		}, nil
	}

	if pet.CurrentStageID < rule.RequiredStageID {
		blockedReason := petDepartureBlockedStageNotReached
		if err := u.repo.Upsert(domain.PetDepartureUpsertInput{
			PetID:         pet.PetID,
			UserID:        pet.UserID,
			RuleID:        rule.ID,
			Status:        petDepartureStatusBlocked,
			BlockedReason: &blockedReason,
			CheckedAt:     checkedAt,
		}); err != nil {
			return RunPetDepartureCheckPetResult{}, err
		}
		return RunPetDepartureCheckPetResult{
			PetID:         string(pet.PetID),
			Status:        petDepartureStatusBlocked,
			BlockedReason: &blockedReason,
		}, nil
	}

	eligibleAt := checkedAt
	if pet.StageReachedAt != nil {
		eligibleAt = pet.StageReachedAt.In(timeutil.LocationJST())
	} else if pet.EligibleAt != nil {
		eligibleAt = pet.EligibleAt.In(timeutil.LocationJST())
	}

	scheduledAt := laterTime(
		minAgeAt,
		eligibleAt.AddDate(0, 0, rule.GraceDaysMax),
	)

	if !checkedAt.Before(scheduledAt) {
		if err := u.repo.Depart(domain.PetDepartureDepartInput{
			PetID:       pet.PetID,
			UserID:      pet.UserID,
			RuleID:      rule.ID,
			EligibleAt:  eligibleAt,
			ScheduledAt: scheduledAt,
			DepartedAt:  checkedAt,
		}); err != nil {
			return RunPetDepartureCheckPetResult{}, err
		}
		return RunPetDepartureCheckPetResult{
			PetID:                string(pet.PetID),
			Status:               petDepartureStatusDeparted,
			EligibleAt:           &eligibleAt,
			ScheduledDepartureAt: &scheduledAt,
			DepartedAt:           &checkedAt,
		}, nil
	}

	if err := u.repo.Upsert(domain.PetDepartureUpsertInput{
		PetID:                pet.PetID,
		UserID:               pet.UserID,
		RuleID:               rule.ID,
		Status:               petDepartureStatusScheduled,
		EligibleAt:           &eligibleAt,
		ScheduledDepartureAt: &scheduledAt,
		CheckedAt:            checkedAt,
	}); err != nil {
		return RunPetDepartureCheckPetResult{}, err
	}

	return RunPetDepartureCheckPetResult{
		PetID:                string(pet.PetID),
		Status:               petDepartureStatusScheduled,
		EligibleAt:           &eligibleAt,
		ScheduledDepartureAt: &scheduledAt,
	}, nil
}

func laterTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
