package presenter

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type FindByPetPresenter struct{}

func NewFindByPetPresenter() *FindByPetPresenter {
	return &FindByPetPresenter{}
}

func (p *FindByPetPresenter) Output(posts []domain.Post) []usecases.FindByPetIDPostOutput {
	var outs []usecases.FindByPetIDPostOutput
	for _, post := range posts {
		emb := ""
		if post.ContentEmbedding() != nil {
			emb = *post.ContentEmbedding()
		}
		outs = append(outs, usecases.FindByPetIDPostOutput{
			ID:               string(post.ID()),
			PetID:            string(post.PetID()),
			Content:          post.Content(),
			ContentEmbedding: emb,
			CreatedAt:        post.CreatedAt(),
		})
	}
	return outs
}
