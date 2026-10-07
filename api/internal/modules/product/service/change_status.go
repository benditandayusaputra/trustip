package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type ChangeProductStatus struct {
	deps       ProductDeps
	isArchived bool
	action     audit.Action
}

func NewArchiveProduct(deps ProductDeps) *ChangeProductStatus {
	return &ChangeProductStatus{deps: deps, isArchived: true, action: audit.ProductArchived}
}

func NewReactivateProduct(deps ProductDeps) *ChangeProductStatus {
	return &ChangeProductStatus{deps: deps, isArchived: false, action: audit.ProductReactivated}
}

func (uc *ChangeProductStatus) Execute(
	ctx context.Context, signedIn actor.Actor, productID string,
) (domain.Product, error) {
	if err := policy.Require(signedIn, policy.ManageProducts); err != nil {
		return domain.Product{}, fmt.Errorf("check product manager: %w", err)
	}
	event := productEvent{signedIn: signedIn, productID: productID, action: uc.action, at: uc.deps.Clock.Now()}
	var product domain.Product
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		key := keyOf(signedIn, productID)
		if _, err := uc.deps.Products.GetProduct(ctx, key); err != nil {
			return fmt.Errorf("find product: %w", err)
		}
		if err := uc.deps.Products.SetArchived(ctx, domain.StatusChange{
			ProductID: productID, ConsignorID: key.ConsignorID, IsArchived: uc.isArchived, At: event.at,
		}); err != nil {
			return fmt.Errorf("set archived: %w", err)
		}
		var err error
		product, err = finishProductEvent(ctx, uc.deps, event)
		return err
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("change status of product %s: %w", productID, err)
	}
	return product, nil
}
