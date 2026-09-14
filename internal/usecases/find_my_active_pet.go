package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

// 認証済みユーザーのアクティブペット取得条件です。
type FindMyActivePetInput struct {
	UserID domain.UserID
}

// ペットが現在いる群れの表示用情報です。
type CurrentGroupOutput struct {
	ID          int
	GroupKey    string
	DisplayName string
}

type DepartureOutput struct {
	Status               string
	EligibleAt           *time.Time
	ScheduledDepartureAt *time.Time
	CanDepart            bool
}

// ホーム画面などで必要な現在のペット情報です。
type FindMyActivePetOutput struct {
	ID             string
	Name           string
	Color          string
	CurrentStageID int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CurrentGroup   *CurrentGroupOutput
	Departure      *DepartureOutput
}

type FindMyActivePet struct {
	petRepo       domain.PetRepository
	groupRepo     domain.GroupMasterRepository
	departureRepo domain.PetDepartureRepository
}

func NewFindMyActivePet(
	petRepo domain.PetRepository,
	groupRepo domain.GroupMasterRepository,
	departureRepo domain.PetDepartureRepository,
) *FindMyActivePet {
	return &FindMyActivePet{
		petRepo:       petRepo,
		groupRepo:     groupRepo,
		departureRepo: departureRepo,
	}
}

func (u *FindMyActivePet) Execute(ctx context.Context, input FindMyActivePetInput) (FindMyActivePetOutput, error) {
	if !domain.IsValidUserID(input.UserID) {
		return FindMyActivePetOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindActiveByUserID(ctx, input.UserID)
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

	if u.departureRepo != nil {
		departure, err := u.departureRepo.FindByPetID(ctx, pet.ID())
		if err == nil {
			output.Departure = &DepartureOutput{
				Status:               departure.Status,
				EligibleAt:           departure.EligibleAt,
				ScheduledDepartureAt: departure.ScheduledDepartureAt,
				CanDepart: (departure.Status == PetDepartureStatusEligible || departure.Status == petDepartureStatusScheduled) &&
					departure.ScheduledDepartureAt != nil &&
					!timeutil.NowJST().Before(departure.ScheduledDepartureAt.In(timeutil.LocationJST())),
			}
		} else if !errors.Is(err, domain.ErrNotFound) {
			return FindMyActivePetOutput{}, err
		}
	}

	// ペットを作成した直後など、まだ所属群れがない状態はnullを返す
	if pet.CurrentGroupMasterID() == nil {
		return output, nil
	}

	group, err := u.groupRepo.FindByID(ctx, domain.GroupMasterID(*pet.CurrentGroupMasterID()))
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
