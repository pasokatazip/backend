package presenter

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CreatePostPresenter struct{}

func NewCreatePostPresenter() *CreatePostPresenter {
	return &CreatePostPresenter{}
}

func (p *CreatePostPresenter) Output(post domain.Post) usecases.CreatePostOutput {
	return usecases.CreatePostOutput{
		ID:               string(post.ID()),
		PetID:            string(post.PetID()),
		Content:          string(post.Content()),
		ContentEmbedding: string(post.ContentEmbedding()),
		CreatedAt:        post.CreatedAt(),
	}
}
