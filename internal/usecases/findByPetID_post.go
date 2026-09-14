package usecases

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindByPetIDPostInput struct {
	UserID domain.UserID
	PetID  domain.PetID
}

type FindByPetIDPost struct {
	postRepo domain.PostRepository
	petRepo  domain.PetRepository
}

func NewFindByPetIDPost(postRepo domain.PostRepository, petRepo domain.PetRepository) *FindByPetIDPost {
	return &FindByPetIDPost{postRepo: postRepo, petRepo: petRepo}
}

func (r *FindByPetIDPost) Execute(ctx context.Context, input FindByPetIDPostInput) ([]domain.Post, error) {
	if _, err := findOwnedPet(ctx, r.petRepo, input.UserID, input.PetID); err != nil {
		return nil, err
	}

	return r.postRepo.FindByPetID(ctx, input.PetID)
}
