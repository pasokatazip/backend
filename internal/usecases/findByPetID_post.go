package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindByPetIDPostInput struct {
	UserID domain.UserID
	PetID  domain.PetID
}

type FindByPetIDPostOutput struct {
	ID               string
	PetID            string
	Content          string
	ContentEmbedding string
	CreatedAt        time.Time
}

type FindByPetIDPost struct {
	postRepo domain.PostRepository
	petRepo  domain.PetRepository
}

func NewFindByPetIDPost(postRepo domain.PostRepository, petRepo domain.PetRepository) *FindByPetIDPost {
	return &FindByPetIDPost{postRepo: postRepo, petRepo: petRepo}
}

func (r *FindByPetIDPost) Execute(input FindByPetIDPostInput) ([]domain.Post, error) {
	if _, err := findOwnedPet(r.petRepo, input.UserID, input.PetID); err != nil {
		return nil, err
	}

	return r.postRepo.FindByPetID(input.PetID)
}
