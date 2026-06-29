package domain

import "time"

type PetID string

type Pet struct {
	id                   PetID
	name                 string
	isDeleted            bool
	userID               UserID
	energy               float64
	curiosity            float64
	sociality            float64
	routine              float64
	currentGroupMasterID *int
	currentStageID       int
	createdAt            time.Time
	updatedAt            time.Time
}

func NewPet(
	id PetID,
	name string,
	isDeleted bool,
	userID UserID,
	energy float64,
	curiosity float64,
	sociality float64,
	routine float64,
	currentGroupMasterID *int,
	currentStageID int,
	createdAt time.Time,
	updatedAt time.Time,
) Pet {
	return Pet{
		id:                   id,
		name:                 name,
		isDeleted:            isDeleted,
		userID:               userID,
		energy:               energy,
		curiosity:            curiosity,
		sociality:            sociality,
		routine:              routine,
		currentGroupMasterID: currentGroupMasterID,
		currentStageID:       currentStageID,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}
}

func (p Pet) ID() PetID {
	return p.id
}

func (p Pet) Name() string {
	return p.name
}

func (p Pet) IsDeleted() bool {
	return p.isDeleted
}

func (p Pet) UserID() UserID {
	return p.userID
}

func (p Pet) Energy() float64 {
	return p.energy
}

func (p Pet) Curiosity() float64 {
	return p.curiosity
}

func (p Pet) Sociality() float64 {
	return p.sociality
}

func (p Pet) Routine() float64 {
	return p.routine
}

func (p Pet) CurrentGroupMasterID() *int {
	return p.currentGroupMasterID
}

func (p Pet) CurrentStageID() int {
	return p.currentStageID
}

func (p Pet) CreatedAt() time.Time {
	return p.createdAt
}

func (p Pet) UpdatedAt() time.Time {
	return p.updatedAt
}
