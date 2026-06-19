package domain

import "time"

type NounGroupMatch struct {
	id              NounGroupMatchID
	extractedNounID ExtractedNounID
	groupMasterID   GroupMasterID
	keywordScore    float64
	vectorScore     float64
	keywordWeight   float64
	matchScore      float64
	matchReason     *string
	selected        bool
	createdAt       time.Time
}

func NewNounGroupMatch(
	id NounGroupMatchID,
	extractedNounID ExtractedNounID,
	groupMasterID GroupMasterID,
	keywordScore float64,
	vectorScore float64,
	keywordWeight float64,
	matchScore float64,
	matchReason *string,
	selected bool,
	createdAt time.Time,
) NounGroupMatch {
	return NounGroupMatch{
		id:              id,
		extractedNounID: extractedNounID,
		groupMasterID:   groupMasterID,
		keywordScore:    keywordScore,
		vectorScore:     vectorScore,
		keywordWeight:   keywordWeight,
		matchScore:      matchScore,
		matchReason:     matchReason,
		selected:        selected,
		createdAt:       createdAt,
	}
}

func (n NounGroupMatch) ID() NounGroupMatchID {
	return n.id
}

func (n NounGroupMatch) ExtractedNounID() ExtractedNounID {
	return n.extractedNounID
}

func (n NounGroupMatch) GroupMasterID() GroupMasterID {
	return n.groupMasterID
}

func (n NounGroupMatch) KeywordScore() float64 {
	return n.keywordScore
}

func (n NounGroupMatch) VectorScore() float64 {
	return n.vectorScore
}

func (n NounGroupMatch) KeywordWeight() float64 {
	return n.keywordWeight
}

func (n NounGroupMatch) MatchScore() float64 {
	return n.matchScore
}

func (n NounGroupMatch) MatchReason() *string {
	return n.matchReason
}

func (n NounGroupMatch) Selected() bool {
	return n.selected
}

func (n NounGroupMatch) CreatedAt() time.Time {
	return n.createdAt
}

type NounGroupMatchRepository interface {
	Create(nounGroupMatch NounGroupMatch) (NounGroupMatch, error)
	FindByExtractedNounID(extractedNounID ExtractedNounID) ([]NounGroupMatch, error)
	FindSelectedByExtractedNounID(extractedNounID ExtractedNounID) (NounGroupMatch, error)
}
