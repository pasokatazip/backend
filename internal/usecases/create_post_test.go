package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

const (
	testCreatePostUserID     = domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	testCreatePostOtherID    = domain.UserID("9cb459e5-a503-4847-8988-fcceaa64277b")
	testCreatePostPetID      = domain.PetID("265faa23-d586-488a-b159-875d8be74f1f")
	testCreatePostOtherPetID = domain.PetID("1add695f-d572-47f0-a0ab-89099fdd351d")
)

type createPostRepositoryStub struct {
	called bool
	post   domain.Post
	err    error
}

func (r *createPostRepositoryStub) CreateWithFeedExperience(_ context.Context, post domain.Post, amount int) (domain.Post, error) {
	r.called = true
	r.post = post
	if amount != feedExperienceAmount {
		return domain.Post{}, errors.New("unexpected experience amount")
	}
	if r.err != nil {
		return domain.Post{}, r.err
	}
	return post, nil
}

func (r *createPostRepositoryStub) FindByPetID(_ context.Context, _ domain.PetID) ([]domain.Post, error) {
	return nil, nil
}

type createPostPetRepositoryStub struct {
	pet       domain.Pet
	activePet domain.Pet
	findErr   error
	activeErr error
}

func (r *createPostPetRepositoryStub) Create(_ context.Context, _ domain.Pet) (domain.Pet, error) {
	return domain.Pet{}, nil
}
func (r *createPostPetRepositoryStub) FindByID(_ context.Context, _ domain.PetID) (domain.Pet, error) {
	return r.pet, r.findErr
}
func (r *createPostPetRepositoryStub) FindActiveByUserID(_ context.Context, _ domain.UserID) (domain.Pet, error) {
	return r.activePet, r.activeErr
}
func (r *createPostPetRepositoryStub) FindAllByUserID(_ context.Context, _ domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}
func (r *createPostPetRepositoryStub) FindDeletedByUserID(_ context.Context, _ domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}
func (r *createPostPetRepositoryStub) UpdateProfile(_ context.Context, _ domain.PetID, _ domain.UserID, _ string, _ string, _ time.Time) (domain.Pet, error) {
	return domain.Pet{}, nil
}

func newCreatePostTestPet(id domain.PetID, userID domain.UserID, deleted bool) domain.Pet {
	return domain.NewPet(id, "pet", domain.DefaultPetColor, deleted, userID, 0, 0, 0, 0, nil, 1, time.Time{}, time.Time{})
}

func TestCreatePostExecute(t *testing.T) {
	ownedPet := newCreatePostTestPet(testCreatePostPetID, testCreatePostUserID, false)

	tests := []struct {
		name        string
		input       CreatePostInput
		petRepo     *createPostPetRepositoryStub
		wantErr     error
		wantCreate  bool
		wantContent string
	}{
		{
			name:        "creates a post for the authenticated user's active pet",
			input:       CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo:     &createPostPetRepositoryStub{pet: ownedPet, activePet: ownedPet},
			wantCreate:  true,
			wantContent: "hello",
		},
		{
			name:        "accepts one character and trims surrounding whitespace",
			input:       CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "  あ  "},
			petRepo:     &createPostPetRepositoryStub{pet: ownedPet, activePet: ownedPet},
			wantCreate:  true,
			wantContent: "あ",
		},
		{
			name:    "rejects empty content",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: ""},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects whitespace-only content",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: " \n\t "},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects an invalid user ID",
			input:   CreatePostInput{UserID: "invalid", PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects an invalid pet ID",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: "invalid", Content: "hello"},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "returns not found when the pet does not exist",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{findErr: domain.ErrNotFound},
			wantErr: domain.ErrNotFound,
		},
		{
			name:    "rejects a pet owned by another user",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{pet: newCreatePostTestPet(testCreatePostPetID, testCreatePostOtherID, false)},
			wantErr: domain.ErrUnauthorized,
		},
		{
			name:    "rejects a deleted pet",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{pet: newCreatePostTestPet(testCreatePostPetID, testCreatePostUserID, true)},
			wantErr: domain.ErrValidation,
		},
		{
			name:  "rejects an inactive pet",
			input: CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{
				pet:       ownedPet,
				activePet: newCreatePostTestPet(testCreatePostOtherPetID, testCreatePostUserID, false),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:  "rejects posting when the user has no active pet",
			input: CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{
				pet:       ownedPet,
				activeErr: domain.ErrNotFound,
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := &createPostRepositoryStub{}
			usecase := NewCreatePost(postRepo, tt.petRepo)

			post, err := usecase.Execute(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.wantErr)
			}
			if postRepo.called != tt.wantCreate {
				t.Fatalf("CreateWithFeedExperience() called = %v, want %v", postRepo.called, tt.wantCreate)
			}
			if tt.wantCreate && (post.PetID() != tt.input.PetID || post.Content() != tt.wantContent) {
				t.Fatalf("Execute() post = (%s, %q), want (%s, %q)", post.PetID(), post.Content(), tt.input.PetID, tt.wantContent)
			}
		})
	}
}
