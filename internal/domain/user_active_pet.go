package domain

import (
	"context"

	"time"
)

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
	Create(ctx context.Context, userActivePet UserActivePet) (UserActivePet, error)
	FindByUserID(ctx context.Context, userID UserID) (UserActivePet, error)
	FindByPetID(ctx context.Context, petID PetID) (UserActivePet, error)
	ReplaceByUserID(ctx context.Context, userActivePet UserActivePet) (UserActivePet, error)
	DeleteByUserID(ctx context.Context, userID UserID) error
}
