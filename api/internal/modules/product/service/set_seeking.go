package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type SetSeekingCommand struct {
	ProductID string
	IsSeeking bool
}

type SetSeeking struct {
	deps ProductDeps
}

func NewSetSeeking(deps ProductDeps) *SetSeeking {
	return &SetSeeking{deps: deps}
}

func (uc *SetSeeking) Execute(
	ctx context.Context, signedIn actor.Actor, cmd SetSeekingCommand,
) (domain.Product, error) {
	if err := policy.Require(signedIn, policy.ManageProducts); err != nil {
		return domain.Product{}, fmt.Errorf("check product manager: %w", err)
	}
	event := productEvent{
		signedIn: signedIn, productID: cmd.ProductID, action: audit.ProductSeekingChanged, at: uc.deps.Clock.Now(),
	}
	var product domain.Product
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		key := keyOf(signedIn, cmd.ProductID)
		if _, err := uc.deps.Products.GetProduct(ctx, key); err != nil {
			return fmt.Errorf("find product: %w", err)
		}
		if err := uc.deps.Products.SetSeeking(ctx, domain.SeekingChange{
			ProductID: cmd.ProductID, ConsignorID: key.ConsignorID, IsSeeking: cmd.IsSeeking, At: event.at,
		}); err != nil {
			return fmt.Errorf("set seeking: %w", err)
		}
		var err error
		product, err = finishProductEvent(ctx, uc.deps, event)
		return err
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("set seeking of product %s: %w", cmd.ProductID, err)
	}
	return product, nil
}
