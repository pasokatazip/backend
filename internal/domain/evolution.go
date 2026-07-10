package domain

import "time"

type EvolutionStage struct {
	id        EvolutionStageID
	stageKey  string
	stageNo   int
	name      string
	branchKey *string
	imageURL  *string
	createdAt time.Time
	updatedAt time.Time
}

func NewEvolutionStage(
	id EvolutionStageID,
	stageKey string,
	stageNo int,
	name string,
	branchKey *string,
	imageURL *string,
	createdAt time.Time,
	updatedAt time.Time,
) EvolutionStage {
	return EvolutionStage{
		id:        id,
		stageKey:  stageKey,
		stageNo:   stageNo,
		name:      name,
		branchKey: branchKey,
		imageURL:  imageURL,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (e EvolutionStage) ID() EvolutionStageID {
	return e.id
}

func (e EvolutionStage) StageKey() string {
	return e.stageKey
}

func (e EvolutionStage) StageNo() int {
	return e.stageNo
}

func (e EvolutionStage) Name() string {
	return e.name
}

func (e EvolutionStage) BranchKey() *string {
	return e.branchKey
}

func (e EvolutionStage) ImageURL() *string {
	return e.imageURL
}

func (e EvolutionStage) CreatedAt() time.Time {
	return e.createdAt
}

func (e EvolutionStage) UpdatedAt() time.Time {
	return e.updatedAt
}

type EvolutionRule struct {
	id                             EvolutionRuleID
	fromStageID                    EvolutionStageID
	toStageID                      EvolutionStageID
	requiredExperience             int64
	requiredDaysSinceLastEvolution int
	requiredFeedCount              int
	appearancePart                 *string
	createdAt                      time.Time
	updatedAt                      time.Time
}

func NewEvolutionRule(
	id EvolutionRuleID,
	fromStageID EvolutionStageID,
	toStageID EvolutionStageID,
	requiredExperience int64,
	requiredDaysSinceLastEvolution int,
	requiredFeedCount int,
	appearancePart *string,
	createdAt time.Time,
	updatedAt time.Time,
) EvolutionRule {
	return EvolutionRule{
		id:                             id,
		fromStageID:                    fromStageID,
		toStageID:                      toStageID,
		requiredExperience:             requiredExperience,
		requiredDaysSinceLastEvolution: requiredDaysSinceLastEvolution,
		requiredFeedCount:              requiredFeedCount,
		appearancePart:                 appearancePart,
		createdAt:                      createdAt,
		updatedAt:                      updatedAt,
	}
}

func (e EvolutionRule) ID() EvolutionRuleID {
	return e.id
}

func (e EvolutionRule) FromStageID() EvolutionStageID {
	return e.fromStageID
}

func (e EvolutionRule) ToStageID() EvolutionStageID {
	return e.toStageID
}

func (e EvolutionRule) RequiredExperience() int64 {
	return e.requiredExperience
}

func (e EvolutionRule) RequiredDaysSinceLastEvolution() int {
	return e.requiredDaysSinceLastEvolution
}

func (e EvolutionRule) RequiredFeedCount() int {
	return e.requiredFeedCount
}

func (e EvolutionRule) AppearancePart() *string {
	return e.appearancePart
}

func (e EvolutionRule) CreatedAt() time.Time {
	return e.createdAt
}

func (e EvolutionRule) UpdatedAt() time.Time {
	return e.updatedAt
}

type PetEvolution struct {
	id              PetEvolutionID
	petID           PetID
	stageID         EvolutionStageID
	evolutionRuleID *EvolutionRuleID
	primaryStatus   *string
	evolvedAt       time.Time
	createdAt       time.Time
}

func NewPetEvolution(
	id PetEvolutionID,
	petID PetID,
	stageID EvolutionStageID,
	evolutionRuleID *EvolutionRuleID,
	primaryStatus *string,
	evolvedAt time.Time,
	createdAt time.Time,
) PetEvolution {
	return PetEvolution{
		id:              id,
		petID:           petID,
		stageID:         stageID,
		evolutionRuleID: evolutionRuleID,
		primaryStatus:   primaryStatus,
		evolvedAt:       evolvedAt,
		createdAt:       createdAt,
	}
}

func (p PetEvolution) ID() PetEvolutionID {
	return p.id
}

func (p PetEvolution) PetID() PetID {
	return p.petID
}

func (p PetEvolution) StageID() EvolutionStageID {
	return p.stageID
}

func (p PetEvolution) EvolutionRuleID() *EvolutionRuleID {
	return p.evolutionRuleID
}

func (p PetEvolution) PrimaryStatus() *string {
	return p.primaryStatus
}

func (p PetEvolution) EvolvedAt() time.Time {
	return p.evolvedAt
}

func (p PetEvolution) CreatedAt() time.Time {
	return p.createdAt
}

type EvolutionStageRepository interface {
	FindByID(id EvolutionStageID) (EvolutionStage, error)
	FindByStageNo(stageNo int) (EvolutionStage, error)
	FindAll() ([]EvolutionStage, error)
}

type EvolutionRuleRepository interface {
	FindByFromStageID(fromStageID EvolutionStageID) ([]EvolutionRule, error)
}

type PetEvolutionRepository interface {
	Create(petEvolution PetEvolution) (PetEvolution, error)
	FindByPetID(petID PetID) ([]PetEvolution, error)
	FindLatestByPetID(petID PetID) (PetEvolution, error)
}
