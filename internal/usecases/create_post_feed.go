package usecases

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

// addFeedExperience は給餌による経験値を加算し、取得上限によって制限された量を履歴に記録する。
func (p *CreatePost) addFeedExperience(ctx context.Context, post domain.Post) error {
	_, cappedAmount, err := p.experience.AddFeedExperience(ctx, post.PetID(), feedExperienceAmount, post.CreatedAt())
	if err != nil {
		return err
	}
	sourceID := string(post.ID())
	event := domain.NewPetExperienceEvent(domain.NewPetExperienceEventID(), post.PetID(), domain.ExperienceSourceTypeFeed, &sourceID, feedExperienceAmount, cappedAmount, post.CreatedAt(), post.CreatedAt())
	return p.events.CreateInTransaction(ctx, event)
}
