package domain

import (
	"context"

	"time"
)

type PetExperience struct {
	id              PetExperienceID
	petID           PetID
	totalExperience int64
	feedCount       int
	createdAt       time.Time
	updatedAt       time.Time
}

func NewPetExperience(
	id PetExperienceID,
	petID PetID,
	totalExperience int64,
	feedCount int,
	createdAt time.Time,
	updatedAt time.Time,
) PetExperience {
	return PetExperience{
		id:              id,
		petID:           petID,
		totalExperience: totalExperience,
		feedCount:       feedCount,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (p PetExperience) ID() PetExperienceID {
	return p.id
}

func (p PetExperience) PetID() PetID {
	return p.petID
}

func (p PetExperience) TotalExperience() int64 {
	return p.totalExperience
}

func (p PetExperience) FeedCount() int {
	return p.feedCount
}

func (p PetExperience) CreatedAt() time.Time {
	return p.createdAt
}

func (p PetExperience) UpdatedAt() time.Time {
	return p.updatedAt
}

type PetExperienceRepository interface {
	Create(ctx context.Context, petExperience PetExperience) (PetExperience, error)
	FindByPetID(ctx context.Context, petID PetID) (PetExperience, error)
	Update(ctx context.Context, petExperience PetExperience) (PetExperience, error)
	AddFeedExperience(context.Context, PetID, int, time.Time) (int, int, error)
}
