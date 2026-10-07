package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
)

type SealStore interface {
	LockHead(ctx context.Context, partnershipID string) (domain.Head, error)
	Append(ctx context.Context, seal domain.Seal) error
}

type Signer interface {
	Sign(digest []byte) domain.Signature
}

type IDGenerator interface {
	NewID() (string, error)
}

type Recorder struct {
	store  SealStore
	signer Signer
	ids    IDGenerator
}

func NewRecorder(store SealStore, signer Signer, ids IDGenerator) *Recorder {
	return &Recorder{store: store, signer: signer, ids: ids}
}

func (r *Recorder) Seal(ctx context.Context, request domain.Request) (domain.Seal, error) {
	id, err := r.ids.NewID()
	if err != nil {
		return domain.Seal{}, fmt.Errorf("generate seal id: %w", err)
	}
	head, err := r.store.LockHead(ctx, request.PartnershipID)
	if err != nil {
		return domain.Seal{}, fmt.Errorf("lock seal head of %s: %w", request.PartnershipID, err)
	}
	entry, err := domain.NewEntry(request, head)
	if err != nil {
		return domain.Seal{}, fmt.Errorf("prepare seal entry: %w", err)
	}
	seal, err := domain.Sealed(id, entry, r.signer.Sign)
	if err != nil {
		return domain.Seal{}, fmt.Errorf("sign seal entry: %w", err)
	}
	if err := r.store.Append(ctx, seal); err != nil {
		return domain.Seal{}, fmt.Errorf("append seal %d of %s: %w", entry.Seq, request.PartnershipID, err)
	}
	return seal, nil
}
