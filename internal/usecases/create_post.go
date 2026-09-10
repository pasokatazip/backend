package usecases

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type CreatePostInput struct {
	UserID  domain.UserID
	Content string
	PetID   domain.PetID
}

type CreatePostOutput struct {
	ID               string
	PetID            string
	Content          string
	ContentEmbedding string
	CreatedAt        time.Time
}

type CreatePost struct {
	postRepo domain.PostRepository
	petRepo  domain.PetRepository
}

const (
	feedExperienceAmount = 10
	maxPostContentLength  = 100
)

func NewCreatePost(postRepo domain.PostRepository, petRepo domain.PetRepository) *CreatePost {
	return &CreatePost{postRepo: postRepo, petRepo: petRepo}
}

func (p *CreatePost) Execute(input CreatePostInput) (domain.Post, error) {
	if !domain.IsValidUserID(input.UserID) ||
		!domain.IsValidPetID(input.PetID) ||
		input.Content == "" ||
		utf8.RuneCountInString(input.Content) > maxPostContentLength {
		return domain.Post{}, domain.ErrValidation
	}

	pet, err := p.petRepo.FindByID(input.PetID)
	if err != nil {
		return domain.Post{}, err
	}
	if pet.UserID() != input.UserID {
		return domain.Post{}, domain.ErrUnauthorized
	}
	if pet.IsDeleted() {
		return domain.Post{}, domain.ErrValidation
	}

	activePet, err := p.petRepo.FindActiveByUserID(input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Post{}, domain.ErrValidation
		}
		return domain.Post{}, err
	}
	if activePet.ID() != input.PetID || activePet.IsDeleted() {
		return domain.Post{}, domain.ErrValidation
	}

	newPost := domain.NewPost(
		domain.NewPostID(),
		input.Content,
		nil,
		input.PetID,
		timeutil.NowJST(),
	)

	savedPost, err := p.postRepo.CreateWithFeedExperience(newPost, feedExperienceAmount)
	if err != nil {
		return domain.Post{}, err
	}

	return savedPost, nil
}
