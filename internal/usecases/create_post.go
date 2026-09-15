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
	transaction domain.Transaction
	petRepo     domain.PetRepository
	posts       domain.PostRepository
	experience  domain.PetExperienceRepository
	events      domain.PetExperienceEventRepository
	rules       domain.EvolutionRuleRepository
	evolutions  domain.PetEvolutionRepository
}

const (
	feedExperienceAmount = 10
	minPostContentLength = 1
	maxPostContentLength = 100
)

func NewCreatePost(
	transaction domain.Transaction,
	petRepo domain.PetRepository,
	posts domain.PostRepository,
	experience domain.PetExperienceRepository,
	events domain.PetExperienceEventRepository,
	rules domain.EvolutionRuleRepository,
	evolutions domain.PetEvolutionRepository,
) *CreatePost {
	return &CreatePost{
		transaction: transaction,
		petRepo:     petRepo,
		posts:       posts,
		experience:  experience,
		events:      events,
		rules:       rules,
		evolutions:  evolutions,
	}
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

	err = p.transaction.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, err := p.posts.Create(txCtx, newPost); err != nil {
			return err
		}
		if err := p.addFeedExperience(txCtx, newPost); err != nil {
			return err
		}
		return p.evolveAfterFeed(txCtx, newPost)
	})
	if err != nil {
		return domain.Post{}, err
	}
	return newPost, nil
}
