package domain

import "time"

type PetRepository interface {
	Create(pet Pet) (Pet, error)
	FindByID(id PetID) (Pet, error)
	FindActiveByUserID(userID UserID) (Pet, error)
	FindDeletedByUserID(userID UserID) ([]Pet, error)
	UpdateProfile(id PetID, userID UserID, name, color string, updatedAt time.Time) (Pet, error)
}
