package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type CreateProduct struct {
	deps ProductDeps
}

func NewCreateProduct(deps ProductDeps) *CreateProduct {
	return &CreateProduct{deps: deps}
}

func (uc *CreateProduct) Execute(
	ctx context.Context, signedIn actor.Actor, form domain.ProductForm,
) (domain.Product, error) {
	if err := policy.Require(signedIn, policy.ManageProducts); err != nil {
		return domain.Product{}, fmt.Errorf("check product manager: %w", err)
	}
	draft, err := domain.NewDraft(form)
	if err != nil {
		return domain.Product{}, fmt.Errorf("validate product: %w", err)
	}
	productID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.Product{}, fmt.Errorf("generate product id: %w", err)
	}
	event := productEvent{
		signedIn: signedIn, productID: productID, action: audit.ProductCreated, at: uc.deps.Clock.Now(),
	}
	var product domain.Product
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		consignorID := signedIn.Membership.OrganizationID
		sku, err := resolveSKU(ctx, uc.deps.Products, draft, consignorID)
		if err != nil {
			return err
		}
		if err := uc.deps.Products.SaveProduct(ctx, domain.Change{
			ProductID: productID, ConsignorID: consignorID, SKU: sku, Details: draft.Details, At: event.at,
		}); err != nil {
			return explainProductFailure(err)
		}
		product, err = finishProductEvent(ctx, uc.deps, event)
		return err
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("create product: %w", err)
	}
	return product, nil
}
