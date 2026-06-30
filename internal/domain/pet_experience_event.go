package domain

import "time"

type ExperienceSourceType string

const (
	ExperienceSourceTypeFeed        ExperienceSourceType = "feed"
	ExperienceSourceTypeHourlyLog   ExperienceSourceType = "hourly_log"
	ExperienceSourceTypeInteraction ExperienceSourceType = "interaction"
	ExperienceSourceTypeBonus       ExperienceSourceType = "bonus"
)

type PetExperienceEvent struct {
	id             PetExperienceEventID
	petID          PetID
	sourceType     ExperienceSourceType
	sourceID       *string
	amount         int
	cappedAmount   int
	experienceDate time.Time
	createdAt      time.Time
}

func NewPetExperienceEvent(
	id PetExperienceEventID,
	petID PetID,
	sourceType ExperienceSourceType,
	sourceID *string,
	amount int,
	cappedAmount int,
	experienceDate time.Time,
	createdAt time.Time,
) PetExperienceEvent {
	return PetExperienceEvent{
		id:             id,
		petID:          petID,
		sourceType:     sourceType,
		sourceID:       sourceID,
		amount:         amount,
		cappedAmount:   cappedAmount,
		experienceDate: experienceDate,
		createdAt:      createdAt,
	}
}

func (p PetExperienceEvent) ID() PetExperienceEventID {
	return p.id
}

func (p PetExperienceEvent) PetID() PetID {
	return p.petID
}

func (p PetExperienceEvent) SourceType() ExperienceSourceType {
	return p.sourceType
}

func (p PetExperienceEvent) SourceID() *string {
	return p.sourceID
}

func (p PetExperienceEvent) Amount() int {
	return p.amount
}

func (p PetExperienceEvent) CappedAmount() int {
	return p.cappedAmount
}

func (p PetExperienceEvent) ExperienceDate() time.Time {
	return p.experienceDate
}

func (p PetExperienceEvent) CreatedAt() time.Time {
	return p.createdAt
}

type PetExperienceEventRepository interface {
	Create(petExperienceEvent PetExperienceEvent) (PetExperienceEvent, error)
	FindByPetID(petID PetID) ([]PetExperienceEvent, error)
	FindByPetIDAndDate(petID PetID, experienceDate time.Time) ([]PetExperienceEvent, error)
}
