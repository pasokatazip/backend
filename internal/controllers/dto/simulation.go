package dto

import "time"

type RunHourlySimulationPetResultResponse struct {
	PetID           string  `json:"pet_id"`
	PreviousGroupID *int    `json:"previous_group_master_id,omitempty"`
	NextGroupID     int     `json:"next_group_master_id"`
	Moved           bool    `json:"moved"`
	MoveProbability float64 `json:"move_probability"`
	AmbientEvent    string  `json:"ambient_event"`
}

type RunHourlySimulationResponse struct {
	SimulatedAt          time.Time                              `json:"simulated_at"`
	TotalPets            int                                    `json:"total_pets"`
	Processed            int                                    `json:"processed"`
	Skipped              int                                    `json:"skipped"`
	InterestPropagations int                                    `json:"interest_propagations"`
	ReportsCreated       int                                    `json:"reports_created"`
	Results              []RunHourlySimulationPetResultResponse `json:"results"`
}
