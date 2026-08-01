package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type LatestPetSouvenirResponse struct {
	Souvenir *LatestPetSouvenirItemResponse `json:"souvenir"`
}

type LatestPetSouvenirItemResponse struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	ImageURL    string    `json:"imageURL"`
	FoundAt     time.Time `json:"foundAt"`
	Reported    bool      `json:"reported"`
}

func NewLatestPetSouvenirResponse(
	output usecases.FindLatestPetSouvenirOutput,
) LatestPetSouvenirResponse {
	response := LatestPetSouvenirResponse{}
	if output.Souvenir == nil {
		return response
	}

	response.Souvenir = &LatestPetSouvenirItemResponse{
		ID:          output.Souvenir.ID,
		DisplayName: output.Souvenir.DisplayName,
		ImageURL:    output.Souvenir.ImageURL,
		FoundAt:     output.Souvenir.FoundAt,
		Reported:    output.Souvenir.Reported,
	}
	return response
}
