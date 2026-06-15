package domain

import "time"

type GroupMasterID int

type GroupMaster struct {
	id             GroupMasterID
	groupKey       string
	displayName    string
	category       *string
	minPetCount    int
	energyDelta    float64
	curiosityDelta float64
	socialityDelta float64
	routineDelta   float64
	active         bool
	createdAt      time.Time
}

func NewGroupMaster(
	id GroupMasterID,
	groupKey string,
	displayName string,
	category *string,
	minPetCount int,
	energyDelta float64,
	curiosityDelta float64,
	socialityDelta float64,
	routineDelta float64,
	active bool,
	createdAt time.Time,
) GroupMaster {
	return GroupMaster{
		id:             id,
		groupKey:       groupKey,
		displayName:    displayName,
		category:       category,
		minPetCount:    minPetCount,
		energyDelta:    energyDelta,
		curiosityDelta: curiosityDelta,
		socialityDelta: socialityDelta,
		routineDelta:   routineDelta,
		active:         active,
		createdAt:      createdAt,
	}
}

func (g GroupMaster) ID() GroupMasterID {
	return g.id
}

func (g GroupMaster) GroupKey() string {
	return g.groupKey
}

func (g GroupMaster) DisplayName() string {
	return g.displayName
}

func (g GroupMaster) Category() *string {
	return g.category
}

func (g GroupMaster) MinPetCount() int {
	return g.minPetCount
}

func (g GroupMaster) EnergyDelta() float64 {
	return g.energyDelta
}

func (g GroupMaster) CuriosityDelta() float64 {
	return g.curiosityDelta
}

func (g GroupMaster) SocialityDelta() float64 {
	return g.socialityDelta
}

func (g GroupMaster) RoutineDelta() float64 {
	return g.routineDelta
}

func (g GroupMaster) Active() bool {
	return g.active
}

func (g GroupMaster) CreatedAt() time.Time {
	return g.createdAt
}

type GroupMasterRepository interface {
	FindActive() ([]GroupMaster, error)
	FindByID(id GroupMasterID) (GroupMaster, error)
	FindByGroupKey(groupKey string) (GroupMaster, error)
}
