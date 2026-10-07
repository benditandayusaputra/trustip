package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type ProductDeps struct {
	Products ProductStore
	Audit    AuditTrail
	Tx       Transactor
	Clock    Clock
	IDs      IDGenerator
}

func keyOf(signedIn actor.Actor, productID string) domain.ProductKey {
	return domain.ProductKey{ConsignorID: signedIn.Membership.OrganizationID, ProductID: productID}
}

func resolveSKU(ctx context.Context, writer ProductWriter, draft domain.Draft, consignorID string) (string, error) {
	if !draft.HasAutoSKU() {
		return draft.SKU, nil
	}
	prefix := domain.SKUPrefix(draft.Name)
	next, err := writer.NextSKUNumber(ctx, domain.SKUSeries{ConsignorID: consignorID, Prefix: prefix})
	if err != nil {
		return "", fmt.Errorf("next sku number for %s: %w", prefix, err)
	}
	return domain.AutoSKU(prefix, next), nil
}

type productEvent struct {
	signedIn  actor.Actor
	productID string
	action    audit.Action
	at        time.Time
}

func recordProductEvent(ctx context.Context, deps ProductDeps, event productEvent) error {
	err := deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: event.signedIn.Membership.OrganizationID, ActorID: event.signedIn.UserID,
		Action: event.action, EntityType: audit.ProductEntity, EntityID: event.productID, At: event.at,
	})
	if err != nil {
		return fmt.Errorf("audit %s: %w", event.action, err)
	}
	return nil
}

func finishProductEvent(ctx context.Context, deps ProductDeps, event productEvent) (domain.Product, error) {
	if err := recordProductEvent(ctx, deps, event); err != nil {
		return domain.Product{}, err
	}
	product, err := deps.Products.GetProduct(ctx, keyOf(event.signedIn, event.productID))
	if err != nil {
		return domain.Product{}, fmt.Errorf("reload product %s: %w", event.productID, err)
	}
	return product, nil
}

func productFields() map[error]apperror.FieldError {
	return map[error]apperror.FieldError{
		domain.ErrSKUTaken:        apperror.InvalidField("sku", apperror.FieldTaken),
		domain.ErrUnknownCategory: apperror.InvalidField("category_id", apperror.FieldUnknownChoice),
		domain.ErrUnknownPhoto:    apperror.InvalidField("photo_file_id", apperror.FieldUnknownChoice),
	}
}

func explainProductFailure(err error) error {
	for cause, problem := range productFields() {
		if errors.Is(err, cause) {
			return apperror.Validation(problem)
		}
	}
	return fmt.Errorf("save product: %w", err)
}
