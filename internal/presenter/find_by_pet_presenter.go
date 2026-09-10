package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
)

type FindByPetPresenter struct{}

func NewFindByPetPresenter() *FindByPetPresenter {
	return &FindByPetPresenter{}
}

func (p *FindByPetPresenter) Output(posts []domain.Post) []dto.PostResponse {
	var outputs []dto.PostResponse
	for _, post := range posts {
		emb := ""
		if post.ContentEmbedding() != nil {
			emb = *post.ContentEmbedding()
		}
		outputs = append(outputs, dto.PostResponse{
			ID:               string(post.ID()),
			PetID:            string(post.PetID()),
			Content:          post.Content(),
			ContentEmbedding: emb,
			CreatedAt:        post.CreatedAt(),
		})
	}
	return outputs
}
