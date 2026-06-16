package domain

import "time"

type ExtractedNoun struct {
	id             ExtractedNounID
	postID         PostID
	nounText       string
	normalizedNoun string
	nounEmbedding  *string
	createdAt      time.Time
}

func NewExtractedNoun(
	id ExtractedNounID,
	postID PostID,
	nounText string,
	normalizedNoun string,
	nounEmbedding *string,
	createdAt time.Time,
) ExtractedNoun {
	return ExtractedNoun{
		id:             id,
		postID:         postID,
		nounText:       nounText,
		normalizedNoun: normalizedNoun,
		nounEmbedding:  nounEmbedding,
		createdAt:      createdAt,
	}
}

func (e ExtractedNoun) ID() ExtractedNounID {
	return e.id
}

func (e ExtractedNoun) PostID() PostID {
	return e.postID
}

func (e ExtractedNoun) NounText() string {
	return e.nounText
}

func (e ExtractedNoun) NormalizedNoun() string {
	return e.normalizedNoun
}

func (e ExtractedNoun) NounEmbedding() *string {
	return e.nounEmbedding
}

func (e ExtractedNoun) CreatedAt() time.Time {
	return e.createdAt
}

type ExtractedNounRepository interface {
	Create(extractedNoun ExtractedNoun) (ExtractedNoun, error)
	FindByPostID(postID PostID) ([]ExtractedNoun, error)
}
