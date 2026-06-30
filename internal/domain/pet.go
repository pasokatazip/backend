package domain

import "time"

type PetID string

const DefaultPetColor = "#FFC1CA"

func IsValidPetColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for _, c := range color[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

type Pet struct {
	id                   PetID
	name                 string
	color                string
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
	color string,
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
		color:                color,
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

func (p Pet) Color() string {
	return p.color
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
