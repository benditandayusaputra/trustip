package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type PublicNeeds interface {
	GetPublicNeed(ctx context.Context, needID string) (domain.PublicNeed, error)
}

type GetPublicNeed struct {
	needs PublicNeeds
}

func NewGetPublicNeed(needs PublicNeeds) *GetPublicNeed {
	return &GetPublicNeed{needs: needs}
}

func (uc *GetPublicNeed) Execute(ctx context.Context, needID string) (domain.PublicNeed, error) {
	need, err := uc.needs.GetPublicNeed(ctx, needID)
	if err != nil {
		return domain.PublicNeed{}, fmt.Errorf("get public need %s: %w", needID, err)
	}
	return need, nil
}
