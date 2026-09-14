package domain

import (
	"context"

	"time"
)

type PetGroupJoinMoveReason string

const (
	PetGroupJoinMoveReasonFeedMatch        PetGroupJoinMoveReason = "feed_match"
	PetGroupJoinMoveReasonHourlySimulation PetGroupJoinMoveReason = "hourly_simulation"
	PetGroupJoinMoveReasonRestNeed         PetGroupJoinMoveReason = "rest_need"
	PetGroupJoinMoveReasonInitial          PetGroupJoinMoveReason = "initial"
)

type PetGroupJoin struct {
	id            PetGroupJoinID
	petID         PetID
	groupMasterID GroupMasterID
	joinedAt      time.Time
	leftAt        *time.Time
	moveReason    *PetGroupJoinMoveReason
	createdAt     time.Time
	updatedAt     time.Time
}

func NewPetGroupJoin(
	id PetGroupJoinID,
	petID PetID,
	groupMasterID GroupMasterID,
	joinedAt time.Time,
	leftAt *time.Time,
	moveReason *PetGroupJoinMoveReason,
	createdAt time.Time,
	updatedAt time.Time,
) PetGroupJoin {
	return PetGroupJoin{
		id:            id,
		petID:         petID,
		groupMasterID: groupMasterID,
		joinedAt:      joinedAt,
		leftAt:        leftAt,
		moveReason:    moveReason,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (p PetGroupJoin) ID() PetGroupJoinID {
	return p.id
}

func (p PetGroupJoin) PetID() PetID {
	return p.petID
}

func (p PetGroupJoin) GroupMasterID() GroupMasterID {
	return p.groupMasterID
}

func (p PetGroupJoin) JoinedAt() time.Time {
	return p.joinedAt
}

func (p PetGroupJoin) LeftAt() *time.Time {
	return p.leftAt
}

func (p PetGroupJoin) MoveReason() *PetGroupJoinMoveReason {
	return p.moveReason
}

func (p PetGroupJoin) CreatedAt() time.Time {
	return p.createdAt
}

func (p PetGroupJoin) UpdatedAt() time.Time {
	return p.updatedAt
}

type PetGroupJoinRepository interface {
	Create(ctx context.Context, petGroupJoin PetGroupJoin) (PetGroupJoin, error)
	FindActiveByPetID(ctx context.Context, petID PetID) (PetGroupJoin, error)
	FindActiveByGroupMasterID(ctx context.Context, groupMasterID GroupMasterID) ([]PetGroupJoin, error)
	FindByPetID(ctx context.Context, petID PetID) ([]PetGroupJoin, error)
	CloseActiveByPetID(ctx context.Context, petID PetID, leftAt time.Time) error
}
