package dto

import "time"

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
