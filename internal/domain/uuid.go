package domain

import (
	"github.com/google/uuid"
)

type UserID string
type PostID string
type ReportID string
type NotificationID string
type ExtractedNounID int
type NounGroupMatchID int
type GroupKeywordID int
type GroupMasterID int

func NewUserID() UserID {
	return UserID(uuid.New().String())
}

func NewPostID() PostID {
	return PostID(uuid.New().String())
}

func NewPetID() PetID {
	return PetID(uuid.New().String())
}

func NewReportID() ReportID {
	return ReportID(uuid.NewString())
}

func NewNotificationID() NotificationID {
	return NotificationID(uuid.NewString())
}

func NewUUIDString() string {
	return uuid.NewString()
}

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func IsValidPetID(id PetID) bool {
	return IsValidUUID(string(id))
}

func IsValidUserID(id UserID) bool {
	return IsValidUUID(string(id))
}

func IsValidPostID(id PostID) bool {
	return IsValidUUID(string(id))
}

func IsValidReportID(id ReportID) bool {
	return IsValidUUID(string(id))
}

func IsValidNotificationID(id NotificationID) bool {
	return IsValidUUID(string(id))
}
