package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CreatePostRequest struct {
	Content string `json:"content"`
	PetID   string `json:"-" swaggerignore:"true"`
}

func (r CreatePostRequest) ToUseCaseInput() usecases.CreatePostInput {
	return usecases.CreatePostInput{
		Content: r.Content,
		PetID:   domain.PetID(r.PetID),
	}
}

type CreatePostResponse struct {
	ID               string    `json:"id"`
	PetID            string    `json:"pet_id"`
	Content          string    `json:"content"`
	ContentEmbedding string    `json:"content_embedding"`
	CreatedAt        time.Time `json:"createdAt"`
}

func NewCreatePostResponse(output usecases.CreatePostOutput) CreatePostResponse {
	return CreatePostResponse{
		ID:               output.ID,
		PetID:            output.PetID,
		Content:          output.Content,
		ContentEmbedding: output.ContentEmbedding,
		CreatedAt:        output.CreatedAt,
	}
}
