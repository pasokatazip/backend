package dto

import "time"

type CurrentGroupResponse struct {
	ID          int    `json:"id"`
	GroupKey    string `json:"group_key"`
	DisplayName string `json:"display_name"`
}

type DepartureResponse struct {
	Status               string     `json:"status"`
	EligibleAt           *time.Time `json:"eligible_at,omitempty"`
	ScheduledDepartureAt *time.Time `json:"scheduled_departure_at,omitempty"`
	CanDepart            bool       `json:"can_depart"`
}

type CurrentPetResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Color          string                `json:"color"`
	CurrentStageID int                   `json:"current_stage_id"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	CurrentGroup   *CurrentGroupResponse `json:"current_group"`
	Departure      *DepartureResponse    `json:"departure"`
}
