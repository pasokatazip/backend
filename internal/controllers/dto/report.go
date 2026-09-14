package dto

import "time"

type ReportsResponse struct {
	Reports    []ReportResponse `json:"reports"`
	HasPraised bool             `json:"hasPraised"`
}

type SubscriptionReportsResponse struct {
	Reports    []ReportResponse      `json:"reports"`
	Pet        SubscriptionReportPet `json:"pet"`
	HasPraised bool                  `json:"hasPraised"`
}

type SubscriptionReportPet struct {
	ID              string    `json:"pet_id"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	CurrentStageKey string    `json:"current_stage_key"`
	CurrentStageNo  int       `json:"current_stage_no"`
	IsDeleted       bool      `json:"is_deleted"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReportResponse struct {
	ID        string             `json:"id"`
	PetID     string             `json:"petID"`
	GroupName string             `json:"groupName"`
	CreatedAt time.Time          `json:"createdAt"`
	Gossip    string             `json:"gossip"`
	HourSlot  int                `json:"hourSlot"`
	Souvenirs []SouvenirResponse `json:"souvenirs"`
	Rumors    []string           `json:"rumors"`
}

type SouvenirResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	ImageURL    string `json:"imageURL"`
}

type HistoricalReportResponse struct {
	ID        string    `json:"ID"`
	PetID     string    `json:"PetID"`
	HourSlot  int       `json:"HourSlot"`
	Gossip    string    `json:"Gossip"`
	GroupName string    `json:"Group_name"`
	CreatedAt time.Time `json:"CreatedAt"`
	Rumors    []string  `json:"rumors"`
}
