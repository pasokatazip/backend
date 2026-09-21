package usecases

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

// evolveAfterFeed は進化条件を満たした場合に現在のステージを更新し、進化履歴を記録する。
func (p *CreatePost) evolveAfterFeed(ctx context.Context, post domain.Post) error {
	rule, err := p.rules.FindSatisfiedAfterFeed(ctx, post.PetID(), post.CreatedAt())
	if err != nil {
		return err
	}
	if rule == nil {
		return nil
	}
	if err := p.petRepo.UpdateEvolutionStage(ctx, post.PetID(), rule.ToStageID, post.CreatedAt()); err != nil {
		return err
	}
	evolution := domain.NewPetEvolution(domain.NewPetEvolutionID(), post.PetID(), rule.ToStageID, &rule.RuleID, &rule.PrimaryStatus, post.CreatedAt(), post.CreatedAt())
	return p.evolutions.CreateInTransaction(ctx, evolution)
}
