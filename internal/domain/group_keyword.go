package domain

import "time"

type GroupKeyword struct {
	id                GroupKeywordID
	groupMasterID     GroupMasterID
	keyword           string
	normalizedKeyword string
	weight            float64
	matchType         string
	active            bool
	createdAt         time.Time
	updatedAt         time.Time
}

func NewGroupKeyword(
	id GroupKeywordID,
	groupMasterID GroupMasterID,
	keyword string,
	normalizedKeyword string,
	weight float64,
	matchType string,
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) GroupKeyword {
	return GroupKeyword{
		id:                id,
		groupMasterID:     groupMasterID,
		keyword:           keyword,
		normalizedKeyword: normalizedKeyword,
		weight:            weight,
		matchType:         matchType,
		active:            active,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (g GroupKeyword) ID() GroupKeywordID {
	return g.id
}

func (g GroupKeyword) GroupMasterID() GroupMasterID {
	return g.groupMasterID
}

func (g GroupKeyword) Keyword() string {
	return g.keyword
}

func (g GroupKeyword) NormalizedKeyword() string {
	return g.normalizedKeyword
}

func (g GroupKeyword) Weight() float64 {
	return g.weight
}

func (g GroupKeyword) MatchType() string {
	return g.matchType
}

func (g GroupKeyword) Active() bool {
	return g.active
}

func (g GroupKeyword) CreatedAt() time.Time {
	return g.createdAt
}

func (g GroupKeyword) UpdatedAt() time.Time {
	return g.updatedAt
}

type GroupKeywordRepository interface {
	FindActive() ([]GroupKeyword, error)
	FindActiveByGroupMasterID(groupMasterID GroupMasterID) ([]GroupKeyword, error)
	FindByNormalizedKeyword(normalizedKeyword string) ([]GroupKeyword, error)
	FindCandidatesByNormalizedNoun(normalizedNoun string) ([]GroupKeyword, error)
}
