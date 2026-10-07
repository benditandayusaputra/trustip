package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type UpdateProductCommand struct {
	ProductID string
	Form      domain.ProductForm
}

type UpdateProduct struct {
	deps ProductDeps
}

func NewUpdateProduct(deps ProductDeps) *UpdateProduct {
	return &UpdateProduct{deps: deps}
}

func (uc *UpdateProduct) Execute(
	ctx context.Context, signedIn actor.Actor, cmd UpdateProductCommand,
) (domain.Product, error) {
	if err := policy.Require(signedIn, policy.ManageProducts); err != nil {
		return domain.Product{}, fmt.Errorf("check product manager: %w", err)
	}
	draft, err := domain.NewDraft(cmd.Form)
	if err != nil {
		return domain.Product{}, fmt.Errorf("validate product: %w", err)
	}
	event := productEvent{
		signedIn: signedIn, productID: cmd.ProductID, action: audit.ProductUpdated, at: uc.deps.Clock.Now(),
	}
	var product domain.Product
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		key := keyOf(signedIn, cmd.ProductID)
		if _, err := uc.deps.Products.GetProduct(ctx, key); err != nil {
			return fmt.Errorf("find product: %w", err)
		}
		sku, err := resolveSKU(ctx, uc.deps.Products, draft, key.ConsignorID)
		if err != nil {
			return err
		}
		if err := uc.deps.Products.SaveProduct(ctx, domain.Change{
			ProductID: key.ProductID, ConsignorID: key.ConsignorID, SKU: sku, Details: draft.Details, At: event.at,
		}); err != nil {
			return explainProductFailure(err)
		}
		product, err = finishProductEvent(ctx, uc.deps, event)
		return err
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("update product %s: %w", cmd.ProductID, err)
	}
	return product, nil
}
