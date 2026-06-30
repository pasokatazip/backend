package domain

import "time"

type UserActivePet struct {
	userID     UserID
	petID      PetID
	assignedAt time.Time
}

func NewUserActivePet(
	userID UserID,
	petID PetID,
	assignedAt time.Time,
) UserActivePet {
	return UserActivePet{
		userID:     userID,
		petID:      petID,
		assignedAt: assignedAt,
	}
}

func (u UserActivePet) UserID() UserID {
	return u.userID
}

func (u UserActivePet) PetID() PetID {
	return u.petID
}

func (u UserActivePet) AssignedAt() time.Time {
	return u.assignedAt
}

type UserActivePetRepository interface {
	Create(userActivePet UserActivePet) (UserActivePet, error)
	FindByUserID(userID UserID) (UserActivePet, error)
	FindByPetID(petID PetID) (UserActivePet, error)
	ReplaceByUserID(userActivePet UserActivePet) (UserActivePet, error)
	DeleteByUserID(userID UserID) error
}
