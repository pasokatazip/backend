package domain

import (
	"time"

	"github.com/pasokatazip/backend/internal/timeutil"
)

type Report struct {
	id                    ReportID
	petID                 PetID
	hourSlot              int
	gossip                string
	groupMasterID         GroupMasterID
	groupName             string
	previousGroupMasterID *GroupMasterID

	moved bool

	behaviorType  string
	behaviorLabel string

	interactionCount int

	energyDelta    int
	curiosityDelta int
	socialityDelta int
	routineDelta   int

	reasonJSON []byte

	createdAt time.Time
	souvenirs []ReportSouvenir
}

type ReportSouvenir struct {
	id          string
	displayName string
	imageURL    string
}

func NewReportSouvenir(id, displayName, imageURL string) ReportSouvenir {
	return ReportSouvenir{id: id, displayName: displayName, imageURL: imageURL}
}

func (s ReportSouvenir) ID() string          { return s.id }
func (s ReportSouvenir) DisplayName() string { return s.displayName }
func (s ReportSouvenir) ImageURL() string    { return s.imageURL }

func NewReport(
	id ReportID,
	petID PetID,
	hourSlot int,
	gossip *string,
	groupMasterID GroupMasterID,
	behaviorType string,
	behaviorLabel string,
) (Report, error) {

	if hourSlot < 0 || hourSlot > 23 {
		return Report{}, ErrValidation
	}

	return Report{
		id:               id,
		petID:            petID,
		hourSlot:         hourSlot,
		gossip:           *gossip,
		groupMasterID:    groupMasterID,
		behaviorType:     behaviorType,
		behaviorLabel:    behaviorLabel,
		moved:            false,
		interactionCount: 0,
		energyDelta:      0,
		curiosityDelta:   0,
		socialityDelta:   0,
		routineDelta:     0,
		createdAt:        timeutil.NowJST(),
	}, nil
}

func NewPersistedReport(
	id ReportID,
	petID PetID,
	hourSlot int,
	gossip *string,
	groupMasterID GroupMasterID,
	behaviorType string,
	behaviorLabel string,
	groupName string,
	createdAt time.Time,
) (Report, error) {
	report, err := NewReport(
		id,
		petID,
		hourSlot,
		gossip,
		groupMasterID,
		behaviorType,
		behaviorLabel,
	)
	if err != nil {
		return Report{}, err
	}
	report.createdAt = createdAt
	report.groupName = groupName
	return report, nil
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
	return r.gossip
}

func (r Report) GroupMasterID() GroupMasterID {
	return r.groupMasterID
}

func (r Report) GroupName() string {
	return r.groupName
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

func (r Report) Souvenirs() []ReportSouvenir {
	return append([]ReportSouvenir(nil), r.souvenirs...)
}

func (r Report) WithSouvenirs(souvenirs []ReportSouvenir) Report {
	r.souvenirs = append([]ReportSouvenir(nil), souvenirs...)
	return r
}

type ReportRepository interface {
	FindByToday(PetID) ([]Report, error)
	FindAllByPetID(PetID) ([]Report, error)
}
