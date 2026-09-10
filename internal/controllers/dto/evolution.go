package dto

import "time"

type EvolutionStageResponse struct {
	ID        int        `json:"id"`
	StageKey  string     `json:"stage_key"`
	StageNo   int        `json:"stage_no"`
	Name      string     `json:"name"`
	BranchKey *string    `json:"branch_key,omitempty"`
	ImageURL  *string    `json:"image_url,omitempty"`
	Unlocked  bool       `json:"unlocked"`
	Current   bool       `json:"current"`
	EvolvedAt *time.Time `json:"evolved_at,omitempty"`
}

type PetGrowthEvolutionResponse struct {
	ID              string    `json:"id"`
	StageID         int       `json:"stage_id"`
	EvolutionRuleID *int      `json:"evolution_rule_id,omitempty"`
	PrimaryStatus   *string   `json:"primary_status,omitempty"`
	EvolvedAt       time.Time `json:"evolved_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type ActivePetEvolutionHistoryResponse struct {
	PetID           string                       `json:"pet_id"`
	CreatedAt       time.Time                    `json:"created_at"`
	CurrentStageKey string                       `json:"current_stage_key"`
	Stages          []EvolutionStageResponse     `json:"stages"`
	Evolutions      []PetGrowthEvolutionResponse `json:"evolutions"`
}

type CurrentEvolutionStageResponse struct {
	ID        int     `json:"id"`
	StageKey  string  `json:"stage_key"`
	StageNo   int     `json:"stage_no"`
	Name      string  `json:"name"`
	BranchKey *string `json:"branch_key,omitempty"`
	ImageURL  *string `json:"image_url,omitempty"`
}

type EvolutionProgressResponse struct {
	Current   int64 `json:"current"`
	Required  int64 `json:"required"`
	Remaining int64 `json:"remaining"`
	Met       bool  `json:"met"`
}

type EvolutionRequirementsResponse struct {
	Experience             EvolutionProgressResponse `json:"experience"`
	FeedCount              EvolutionProgressResponse `json:"feed_count"`
	DaysSinceLastEvolution EvolutionProgressResponse `json:"days_since_last_evolution"`
}

type EvolutionCandidateResponse struct {
	RuleID         int                           `json:"rule_id"`
	ToStage        CurrentEvolutionStageResponse `json:"to_stage"`
	SelectedForPet bool                          `json:"selected_for_pet"`
	Ready          bool                          `json:"ready"`
	Requirements   EvolutionRequirementsResponse `json:"requirements"`
}

type CurrentPetEvolutionStatusResponse struct {
	PetID        string                        `json:"pet_id"`
	CurrentStage CurrentEvolutionStageResponse `json:"current_stage"`
	CanEvolve    bool                          `json:"can_evolve"`
	NextStages   []EvolutionCandidateResponse  `json:"next_stages"`
}
