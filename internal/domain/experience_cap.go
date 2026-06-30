package domain

import "time"

type ExperienceCapType string

const (
	ExperienceCapTypeDaily   ExperienceCapType = "daily"
	ExperienceCapTypeWeekly  ExperienceCapType = "weekly"
	ExperienceCapTypeMonthly ExperienceCapType = "monthly"
)

type ExperienceCap struct {
	id            ExperienceCapID
	capType       ExperienceCapType
	maxExperience int
	active        bool
	createdAt     time.Time
	updatedAt     time.Time
}

func NewExperienceCap(
	id ExperienceCapID,
	capType ExperienceCapType,
	maxExperience int,
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) ExperienceCap {
	return ExperienceCap{
		id:            id,
		capType:       capType,
		maxExperience: maxExperience,
		active:        active,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (e ExperienceCap) ID() ExperienceCapID {
	return e.id
}

func (e ExperienceCap) CapType() ExperienceCapType {
	return e.capType
}

func (e ExperienceCap) MaxExperience() int {
	return e.maxExperience
}

func (e ExperienceCap) Active() bool {
	return e.active
}

func (e ExperienceCap) CreatedAt() time.Time {
	return e.createdAt
}

func (e ExperienceCap) UpdatedAt() time.Time {
	return e.updatedAt
}

type ExperienceCapRepository interface {
	FindActiveByCapType(capType ExperienceCapType) (ExperienceCap, error)
	FindActive() ([]ExperienceCap, error)
}
