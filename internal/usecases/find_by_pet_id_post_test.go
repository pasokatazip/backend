package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type findPostRepositoryStub struct {
	domain.PostRepository
	called bool
	posts  []domain.Post
}

func (r *findPostRepositoryStub) FindByPetID(domain.PetID) ([]domain.Post, error) {
	r.called = true
	return r.posts, nil
}

type findPostPetRepositoryStub struct {
	domain.PetRepository
	pet domain.Pet
}

func (r *findPostPetRepositoryStub) FindByID(domain.PetID) (domain.Pet, error) {
	return r.pet, nil
}

func TestFindByPetIDPostChecksOwnershipBeforeFetchingPosts(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	requestUserID := domain.UserID("c9428888-122b-11e1-b85c-61cd3cbb3210")
	ownerID := domain.UserID("a9428888-122b-11e1-b85c-61cd3cbb3210")
	postRepo := &findPostRepositoryStub{}
	petRepo := &findPostPetRepositoryStub{
		pet: domain.NewPet(
			petID, "pet", domain.DefaultPetColor, false, ownerID,
			0, 0, 0, 0, nil, 1, time.Time{}, time.Time{},
		),
	}

	_, err := NewFindByPetIDPost(postRepo, petRepo).Execute(FindByPetIDPostInput{
		UserID: requestUserID,
		PetID:  petID,
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("error = %v, want ErrUnauthorized", err)
	}
	if postRepo.called {
		t.Fatal("post repository was called before ownership validation")
	}
}
