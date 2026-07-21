package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

// 認証済みユーザーのアクティブペット取得条件です。
type FindMyActivePetInput struct {
	UserID domain.UserID
}

// ペットが現在いる群れの表示用情報です。
type CurrentGroupOutput struct {
	ID          int    `json:"id"`
	GroupKey    string `json:"group_key"`
	DisplayName string `json:"display_name"`
}

// ホーム画面などで必要な現在のペット情報です。
type FindMyActivePetOutput struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Color          string              `json:"color"`
	CurrentStageID int                 `json:"current_stage_id"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	CurrentGroup   *CurrentGroupOutput `json:"current_group"`
}

type FindMyActivePet struct {
	petRepo   domain.PetRepository
	groupRepo domain.GroupMasterRepository
}

func NewFindMyActivePet(
	petRepo domain.PetRepository,
	groupRepo domain.GroupMasterRepository,
) *FindMyActivePet {
	return &FindMyActivePet{
		petRepo:   petRepo,
		groupRepo: groupRepo,
	}
}

func (u *FindMyActivePet) Execute(input FindMyActivePetInput) (FindMyActivePetOutput, error) {
	if !domain.IsValidUserID(input.UserID) {
		return FindMyActivePetOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindActiveByUserID(input.UserID)
	if err != nil {
		return FindMyActivePetOutput{}, err
	}

	output := FindMyActivePetOutput{
		ID:             string(pet.ID()),
		Name:           pet.Name(),
		Color:          pet.Color(),
		CurrentStageID: pet.CurrentStageID(),
		CreatedAt:      pet.CreatedAt(),
		UpdatedAt:      pet.UpdatedAt(),
	}

	// ペットを作成した直後など、まだ所属群れがない状態はnullを返す
	if pet.CurrentGroupMasterID() == nil {
		return output, nil
	}

	group, err := u.groupRepo.FindByID(domain.GroupMasterID(*pet.CurrentGroupMasterID()))
	if err != nil {
		return FindMyActivePetOutput{}, err
	}

	output.CurrentGroup = &CurrentGroupOutput{
		ID:          int(group.ID()),
		GroupKey:    group.GroupKey(),
		DisplayName: group.DisplayName(),
	}
	return output, nil
}
