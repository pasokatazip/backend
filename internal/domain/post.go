package domain

import (
	"time"
)

type Post struct {
	id               PostID
	content          string
	contentEmbedding *string
	petID            PetID
	createdAt        time.Time
}

func NewPost(id PostID, content string, contentEmbedding *string, petID PetID, createdAt time.Time) Post {
	return Post{
		id:               id,
		content:          content,
		contentEmbedding: contentEmbedding,
		petID:            petID,
		createdAt:        createdAt,
	}
}

func (p Post) ID() PostID {
	return p.id
}

func (p Post) Content() string {
	return p.content
}

func (p Post) ContentEmbedding() *string {
	return p.contentEmbedding
}

func (p Post) PetID() PetID {
	return p.petID
}

func (p Post) CreatedAt() time.Time {
	return p.createdAt
}

type PostRepository interface {
	Create(post Post) (Post, error)
	FindByPetID(petID PetID) ([]Post, error)
}
