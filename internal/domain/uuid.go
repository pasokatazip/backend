package domain

import (
	"github.com/google/uuid"
)

type UserID string
type PostID string
type ExtractedNounID int
type NounGroupMatchID int
type GroupKeywordID int
type GroupMasterID int

// pet作成前のためいったん仮置き
type PetID string

func NewUserID() UserID {
	return UserID(uuid.New().String())
}

func NewPostID() PostID {
	return PostID(uuid.New().String())
}

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func IsValidUserID(id UserID) bool {
	return IsValidUUID(string(id))
}

func IsValidPostID(id PostID) bool {
	return IsValidUUID(string(id))
}
