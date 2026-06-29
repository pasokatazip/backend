package domain

import "time"

type PetHourlyLogID string

type PetHourlyLog struct {
	id                       PetHourlyLogID
	petID                    PetID
	groupMasterID            GroupMasterID
	petGroupJoinID           *string
	simulatedAt              time.Time
	stayed                   bool
	moveProbability          float64
	boredom                  float64
	restNeed                 float64
	currentGroupFit          float64
	attachmentToCurrentGroup float64
	recentMovePenalty        float64
	energyDeltaApplied       float64
	curiosityDeltaApplied    float64
	socialityDeltaApplied    float64
	routineDeltaApplied      float64
	interactionCount         int
	ambientEvent             *string
	reportMaterial           *string
	createdAt                time.Time
}

func NewPetHourlyLog(
	id PetHourlyLogID,
	petID PetID,
	groupMasterID GroupMasterID,
	petGroupJoinID *string,
	simulatedAt time.Time,
	stayed bool,
	moveProbability float64,
	boredom float64,
	restNeed float64,
	currentGroupFit float64,
	attachmentToCurrentGroup float64,
	recentMovePenalty float64,
	energyDeltaApplied float64,
	curiosityDeltaApplied float64,
	socialityDeltaApplied float64,
	routineDeltaApplied float64,
	interactionCount int,
	ambientEvent *string,
	reportMaterial *string,
	createdAt time.Time,
) PetHourlyLog {
	return PetHourlyLog{
		id:                       id,
		petID:                    petID,
		groupMasterID:            groupMasterID,
		petGroupJoinID:           petGroupJoinID,
		simulatedAt:              simulatedAt,
		stayed:                   stayed,
		moveProbability:          moveProbability,
		boredom:                  boredom,
		restNeed:                 restNeed,
		currentGroupFit:          currentGroupFit,
		attachmentToCurrentGroup: attachmentToCurrentGroup,
		recentMovePenalty:        recentMovePenalty,
		energyDeltaApplied:       energyDeltaApplied,
		curiosityDeltaApplied:    curiosityDeltaApplied,
		socialityDeltaApplied:    socialityDeltaApplied,
		routineDeltaApplied:      routineDeltaApplied,
		interactionCount:         interactionCount,
		ambientEvent:             ambientEvent,
		reportMaterial:           reportMaterial,
		createdAt:                createdAt,
	}
}

func NewPetHourlyLogID() PetHourlyLogID {
	return PetHourlyLogID(NewUUIDString())
}

func (l PetHourlyLog) ID() PetHourlyLogID                { return l.id }
func (l PetHourlyLog) PetID() PetID                      { return l.petID }
func (l PetHourlyLog) GroupMasterID() GroupMasterID      { return l.groupMasterID }
func (l PetHourlyLog) PetGroupJoinID() *string           { return l.petGroupJoinID }
func (l PetHourlyLog) SimulatedAt() time.Time            { return l.simulatedAt }
func (l PetHourlyLog) Stayed() bool                      { return l.stayed }
func (l PetHourlyLog) MoveProbability() float64          { return l.moveProbability }
func (l PetHourlyLog) Boredom() float64                  { return l.boredom }
func (l PetHourlyLog) RestNeed() float64                 { return l.restNeed }
func (l PetHourlyLog) CurrentGroupFit() float64          { return l.currentGroupFit }
func (l PetHourlyLog) AttachmentToCurrentGroup() float64 { return l.attachmentToCurrentGroup }
func (l PetHourlyLog) RecentMovePenalty() float64        { return l.recentMovePenalty }
func (l PetHourlyLog) EnergyDeltaApplied() float64       { return l.energyDeltaApplied }
func (l PetHourlyLog) CuriosityDeltaApplied() float64    { return l.curiosityDeltaApplied }
func (l PetHourlyLog) SocialityDeltaApplied() float64    { return l.socialityDeltaApplied }
func (l PetHourlyLog) RoutineDeltaApplied() float64      { return l.routineDeltaApplied }
func (l PetHourlyLog) InteractionCount() int             { return l.interactionCount }
func (l PetHourlyLog) AmbientEvent() *string             { return l.ambientEvent }
func (l PetHourlyLog) ReportMaterial() *string           { return l.reportMaterial }
func (l PetHourlyLog) CreatedAt() time.Time              { return l.createdAt }
