package domain

import "time"

type PetSouvenir struct {
	id          string
	displayName string
	imageURL    string
	foundAt     time.Time
	reported    bool
}

func NewPetSouvenir(
	id string,
	displayName string,
	imageURL string,
	foundAt time.Time,
	reported bool,
) PetSouvenir {
	return PetSouvenir{
		id:          id,
		displayName: displayName,
		imageURL:    imageURL,
		foundAt:     foundAt,
		reported:    reported,
	}
}

func (s PetSouvenir) ID() string          { return s.id }
func (s PetSouvenir) DisplayName() string { return s.displayName }
func (s PetSouvenir) ImageURL() string    { return s.imageURL }
func (s PetSouvenir) FoundAt() time.Time  { return s.foundAt }
func (s PetSouvenir) Reported() bool      { return s.reported }

type PetSouvenirRepository interface {
	// FindLatestByActivePetUserID returns ErrNotFound when the user has no
	// active pet. A nil souvenir with a nil error means the active pet has not
	// found a souvenir yet.
	FindLatestByActivePetUserID(userID UserID) (*PetSouvenir, error)

	// FindLatestByHistoricalPetID returns ErrNotFound when the pet is not a
	// historical pet owned by the user. A nil souvenir with a nil error means
	// the historical pet did not find a souvenir.
	FindLatestByHistoricalPetID(userID UserID, petID PetID) (*PetSouvenir, error)
}
