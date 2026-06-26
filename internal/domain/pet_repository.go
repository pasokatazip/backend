package domain

type PetRepository interface {
	Create(pet Pet) (Pet, error)
	FindByID(id PetID) (Pet, error)
	FindActiveByUserID(userID UserID) (Pet, error)
	FindDeletedByUserID(userID UserID) ([]Pet, error)
}
