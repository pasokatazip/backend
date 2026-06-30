package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type CreatePostInput struct {
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
	repo domain.PostRepository
}

const feedExperienceAmount = 10

func NewCreatePost(repo domain.PostRepository) *CreatePost {
	return &CreatePost{repo: repo}
}

func (p *CreatePost) Execute(input CreatePostInput) (domain.Post, error) {
	if input.Content == "" {
		return domain.Post{}, domain.ErrValidation
	}

	newPost := domain.NewPost(
		domain.NewPostID(),
		input.Content,
		nil,
		input.PetID,
		timeutil.NowJST(),
	)

	savedPost, err := p.repo.CreateWithFeedExperience(newPost, feedExperienceAmount)
	if err != nil {
		return domain.Post{}, err
	}

	return savedPost, nil
}
