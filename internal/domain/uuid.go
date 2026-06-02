package domain

import (
	"github.com/google/uuid"
)

type UserID string

func NewUserID() UserID {
	return UserID(uuid.New().String())
}

func IsValidUserID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}