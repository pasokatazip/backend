package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

const (
	PetDepartureStatusEligible = "eligible"
	PetDepartureStatusDeparted = "departed"
)

type UpdatePetDepartureStatus struct {
	repo domain.PetDepartureRepository
}

type UpdatePetDepartureStatusInput struct {
	UserID    domain.UserID
	Status    string
	CheckedAt *time.Time
}

type UpdatePetDepartureStatusOutput struct {
	PetID                string     `json:"pet_id"`
	Status               string     `json:"status"`
	EligibleAt           time.Time  `json:"eligible_at"`
	ScheduledDepartureAt time.Time  `json:"scheduled_departure_at"`
	DepartedAt           *time.Time `json:"departed_at,omitempty"`
}

func NewUpdatePetDepartureStatus(repo domain.PetDepartureRepository) *UpdatePetDepartureStatus {
	return &UpdatePetDepartureStatus{repo: repo}
}

func (u *UpdatePetDepartureStatus) Execute(input UpdatePetDepartureStatusInput) (UpdatePetDepartureStatusOutput, error) {
	if !domain.IsValidUserID(input.UserID) || (input.Status != PetDepartureStatusEligible && input.Status != PetDepartureStatusDeparted) {
		return UpdatePetDepartureStatusOutput{}, domain.ErrValidation
	}

	checkedAt := timeutil.NowJST()
	if input.CheckedAt != nil {
		checkedAt = input.CheckedAt.In(timeutil.LocationJST())
	}

	rule, err := u.repo.FindActiveRule()
	if err != nil {
		return UpdatePetDepartureStatusOutput{}, err
	}

	pets, err := u.repo.FindActivePetsByUserID(rule, input.UserID)
	if err != nil {
		return UpdatePetDepartureStatusOutput{}, err
	}
	if len(pets) == 0 {
		return UpdatePetDepartureStatusOutput{}, domain.ErrNotFound
	}

	pet := pets[len(pets)-1]
	minAgeAt := pet.CreatedAt.In(timeutil.LocationJST()).AddDate(0, 0, rule.MinAgeDays)
	if checkedAt.Before(minAgeAt) || pet.CurrentStageNo < rule.RequiredStageNo {
		return UpdatePetDepartureStatusOutput{}, domain.ErrValidation
	}

	eligibleAt := checkedAt
	if pet.StageReachedAt != nil {
		eligibleAt = pet.StageReachedAt.In(timeutil.LocationJST())
	} else if pet.EligibleAt != nil {
		eligibleAt = pet.EligibleAt.In(timeutil.LocationJST())
	}
	scheduledAt := laterTime(minAgeAt, eligibleAt.AddDate(0, 0, rule.GraceDaysMax))

	output := UpdatePetDepartureStatusOutput{
		PetID:                string(pet.PetID),
		Status:               input.Status,
		EligibleAt:           eligibleAt,
		ScheduledDepartureAt: scheduledAt,
	}

	if input.Status == PetDepartureStatusEligible {
		if err := u.repo.Upsert(domain.PetDepartureUpsertInput{
			PetID:                pet.PetID,
			UserID:               pet.UserID,
			RuleID:               rule.ID,
			Status:               PetDepartureStatusEligible,
			EligibleAt:           &eligibleAt,
			ScheduledDepartureAt: &scheduledAt,
			CheckedAt:            checkedAt,
		}); err != nil {
			return UpdatePetDepartureStatusOutput{}, err
		}
		return output, nil
	}

	if checkedAt.Before(scheduledAt) {
		return UpdatePetDepartureStatusOutput{}, domain.ErrValidation
	}
	if err := u.repo.Depart(domain.PetDepartureDepartInput{
		PetID:       pet.PetID,
		UserID:      pet.UserID,
		RuleID:      rule.ID,
		EligibleAt:  eligibleAt,
		ScheduledAt: scheduledAt,
		DepartedAt:  checkedAt,
	}); err != nil {
		return UpdatePetDepartureStatusOutput{}, err
	}

	output.DepartedAt = &checkedAt
	return output, nil
}
