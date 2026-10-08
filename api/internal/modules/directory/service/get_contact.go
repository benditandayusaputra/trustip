package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ContactReader interface {
	GetContact(ctx context.Context, organizationID string) (domain.Contact, error)
}

type GetContact struct {
	contacts ContactReader
}

func NewGetContact(contacts ContactReader) *GetContact {
	return &GetContact{contacts: contacts}
}

func (uc *GetContact) Execute(
	ctx context.Context, signedIn actor.Actor, organizationID string,
) (domain.Contact, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.Contact{}, fmt.Errorf("check contact reader: %w", err)
	}
	contact, err := uc.contacts.GetContact(ctx, organizationID)
	if err != nil {
		return domain.Contact{}, fmt.Errorf("get contact of %s: %w", organizationID, err)
	}
	return contact, nil
}
