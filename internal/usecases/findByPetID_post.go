package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindByPetIDPostInput struct {
	PetID domain.PetID
}

type FindByPetIDPostOutput struct {
	ID               string
	PetID            string
	Content          string
	ContentEmbedding string
	CreatedAt        time.Time
}

type FindByPetIDPost struct {
	repo domain.PostRepository
}

func NewFindByPetIDPost(repo domain.PostRepository) *FindByPetIDPost {
	return &FindByPetIDPost{repo: repo}
}

func (r *FindByPetIDPost) Execute(input FindByPetIDPostInput) ([]FindByPetIDPostOutput, error) {
	if input.PetID == "" || !domain.IsValidPetID(input.PetID) {
		return nil, domain.ErrValidation
	}

	posts, err := r.repo.FindByPetID(input.PetID)
	if err != nil {
		return nil, err
	}

	var outputs []FindByPetIDPostOutput
	for _, post := range posts {
		emb := ""
		if post.ContentEmbedding() != nil {
			emb = *post.ContentEmbedding()
		}
		outputs = append(outputs, FindByPetIDPostOutput{
			ID:               string(post.ID()),
			PetID:            string(post.PetID()),
			Content:          post.Content(),
			ContentEmbedding: emb,
			CreatedAt:        post.CreatedAt(),
		})
	}

	return outputs, nil
}
