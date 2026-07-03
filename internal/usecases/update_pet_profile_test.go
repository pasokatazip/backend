package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type updatePetProfileRepository struct {
	petID  domain.PetID
	userID domain.UserID
	name   string
	color  string
}

func (r *updatePetProfileRepository) Create(pet domain.Pet) (domain.Pet, error) {
	return pet, nil
}

func (r *updatePetProfileRepository) FindByID(domain.PetID) (domain.Pet, error) {
	return domain.Pet{}, nil
}

func (r *updatePetProfileRepository) FindActiveByUserID(domain.UserID) (domain.Pet, error) {
	return domain.Pet{}, nil
}

func (r *updatePetProfileRepository) FindDeletedByUserID(domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}

func (r *updatePetProfileRepository) UpdateProfile(
	petID domain.PetID,
	userID domain.UserID,
	name string,
	color string,
	updatedAt time.Time,
) (domain.Pet, error) {
	r.petID = petID
	r.userID = userID
	r.name = name
	r.color = color
	return domain.NewPet(petID, name, color, false, userID, 0, 0, 0, 0, nil, 0, updatedAt, updatedAt), nil
}

func TestUpdatePetProfile(t *testing.T) {
	repo := &updatePetProfileRepository{}
	usecase := NewUpdatePetProfile(repo)
	input := UpdatePetProfileInput{
		PetID:  "d9428888-122b-11e1-b85c-61cd3cbb3210",
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Name:   "ミケ",
		Color:  "#12AbEF",
	}

	pet, err := usecase.Execute(input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if repo.petID != input.PetID || repo.userID != input.UserID {
		t.Fatalf("update scope = (%q, %q)", repo.petID, repo.userID)
	}
	if pet.Name() != input.Name || pet.Color() != input.Color {
		t.Fatalf("updated pet = (%q, %q)", pet.Name(), pet.Color())
	}
}

func TestUpdatePetProfileRejectsInvalidInput(t *testing.T) {
	valid := UpdatePetProfileInput{
		PetID:  "d9428888-122b-11e1-b85c-61cd3cbb3210",
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Name:   "ミケ",
		Color:  "#12ABEF",
	}
	tests := []struct {
		name  string
		input UpdatePetProfileInput
	}{
		{name: "invalid pet ID", input: func() UpdatePetProfileInput { v := valid; v.PetID = "invalid"; return v }()},
		{name: "invalid user ID", input: func() UpdatePetProfileInput { v := valid; v.UserID = "invalid"; return v }()},
		{name: "empty name", input: func() UpdatePetProfileInput { v := valid; v.Name = ""; return v }()},
		{name: "invalid color", input: func() UpdatePetProfileInput { v := valid; v.Color = "pink"; return v }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUpdatePetProfile(&updatePetProfileRepository{}).Execute(tt.input)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("Execute() error = %v, want ErrValidation", err)
			}
		})
	}
}
