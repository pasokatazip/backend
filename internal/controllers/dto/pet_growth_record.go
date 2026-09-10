package dto

import "time"

type PetGrowthExperienceEventResponse struct {
	ID             string    `json:"id"`
	SourceType     string    `json:"source_type"`
	SourceID       *string   `json:"source_id,omitempty"`
	Amount         int       `json:"amount"`
	CappedAmount   int       `json:"capped_amount"`
	ExperienceDate time.Time `json:"experience_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type PetGrowthRecordResponse struct {
	PetID            string                             `json:"pet_id"`
	Color            string                             `json:"color"`
	CreatedAt        time.Time                          `json:"created_at"`
	CurrentStageID   int                                `json:"current_stage_id"`
	Stages           []EvolutionStageResponse           `json:"stages"`
	TotalExperience  int64                              `json:"total_experience"`
	FeedCount        int                                `json:"feed_count"`
	ExperienceEvents []PetGrowthExperienceEventResponse `json:"experience_events"`
	Evolutions       []PetGrowthEvolutionResponse       `json:"evolutions"`
}
