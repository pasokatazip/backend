package usecases

import (
	"context"
	"errors"
	"strings"
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
	minPostContentLength = 1
	maxPostContentLength = 100
)

func NewCreatePost(postRepo domain.PostRepository, petRepo domain.PetRepository) *CreatePost {
	return &CreatePost{postRepo: postRepo, petRepo: petRepo}
}

func (p *CreatePost) Execute(ctx context.Context, input CreatePostInput) (domain.Post, error) {
	content := strings.TrimSpace(input.Content)
	contentLength := utf8.RuneCountInString(content)
	if contentLength < minPostContentLength || contentLength > maxPostContentLength {
		return domain.Post{}, domain.ErrValidation
	}

	pet, err := findOwnedPet(ctx, p.petRepo, input.UserID, input.PetID)
	if err != nil {
		return domain.Post{}, err
	}
	if pet.IsDeleted() {
		return domain.Post{}, domain.ErrValidation
	}

	activePet, err := p.petRepo.FindActiveByUserID(ctx, input.UserID)
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
		content,
		nil,
		input.PetID,
		timeutil.NowJST(),
	)

	savedPost, err := p.postRepo.CreateWithFeedExperience(ctx, newPost, feedExperienceAmount)
	if err != nil {
		return domain.Post{}, err
	}

	return savedPost, nil
}
