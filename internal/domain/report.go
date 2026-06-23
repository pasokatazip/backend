package domain

import(
	"time"
)

type Report struct {
	id                     ReportID
	petID                  PetID
	hourSlot               int
	gossip				   string
	groupMasterID          GroupMasterID
	previousGroupMasterID  *GroupMasterID

	moved                  bool

	behaviorType           string
	behaviorLabel          string

	interactionCount       int

	energyDelta            int
	curiosityDelta         int
	socialityDelta         int
	routineDelta           int

	reasonJSON             []byte 

	createdAt              time.Time
}

func NewReport(
	id ReportID,
	petID PetID,
	hourSlot int,
	gossip   *string,
	groupMasterID GroupMasterID,
	behaviorType string,
	behaviorLabel string,
) (Report,error) {

	if hourSlot < 0 || hourSlot > 23 {
		return Report{}, ErrValidation
	} 

	return Report{
		id:               id,
		petID:            petID,
		hourSlot:         hourSlot,
		gossip: 		  *gossip,
		groupMasterID:    groupMasterID,
		behaviorType:     behaviorType,
		behaviorLabel:    behaviorLabel,
		moved:            false,
		interactionCount: 0,
		energyDelta:      0,
		curiosityDelta:   0,
		socialityDelta:   0,
		routineDelta:     0,
		createdAt:        time.Now().UTC(),
	},nil
}

func (r Report) ID() ReportID {
	return r.id
}

func (r Report) PetID() PetID {
	return r.petID
}

func (r Report) HourSlot() int {
	return r.hourSlot
}

func (r Report) Gossip() string {
	return  r.gossip
}

func (r Report) GroupMasterID() GroupMasterID {
	return r.groupMasterID
}

func (r Report) PreviousGroupMasterID() *GroupMasterID {
	return r.previousGroupMasterID
}

func (r Report) Moved() bool {
	return r.moved
}

func (r Report) BehaviorType() string {
	return r.behaviorType
}

func (r Report) BehaviorLabel() string {
	return r.behaviorLabel
}

func (r Report) InteractionCount() int {
	return r.interactionCount
}

func (r Report) EnergyDelta() int {
	return r.energyDelta
}

func (r Report) CuriosityDelta() int {
	return r.curiosityDelta
}

func (r Report) SocialityDelta() int {
	return r.socialityDelta
}

func (r Report) RoutineDelta() int {
	return r.routineDelta
}

func (r Report) ReasonJSON() []byte {
	return r.reasonJSON
}

func (r Report) CreatedAt() time.Time {
	return r.createdAt
}

type ReportRepository interface {
	FindByToday(PetID) ([]Report, error)
}
