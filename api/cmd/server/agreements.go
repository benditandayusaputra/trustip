package main

import (
	"context"
	"errors"
	"fmt"

	agreementdomain "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	agreementrepository "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/repository"
	agreementservice "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/service"
	proposaldomain "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	sealdomain "github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	sealrepository "github.com/benditandayusaputra/trustip/api/internal/modules/seal/repository"
	sealservice "github.com/benditandayusaputra/trustip/api/internal/modules/seal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

type sealSigner struct {
	sealer *security.Sealer
}

func (s sealSigner) Sign(digest []byte) sealdomain.Signature {
	signature := s.sealer.Sign(digest)
	return sealdomain.Signature{KeyID: signature.KeyID, Bytes: signature.Bytes}
}

func newSealRecorder(pool *database.Pool, keys security.ActiveKeyring) *sealservice.Recorder {
	return sealservice.NewRecorder(sealrepository.NewSeals(pool), sealSigner{sealer: security.NewSealer(keys)},
		idgen.UUIDv7{})
}

type agreementSealer struct {
	recorder *sealservice.Recorder
}

func (s agreementSealer) Seal(ctx context.Context, request agreementdomain.SealRequest) error {
	_, err := s.recorder.Seal(ctx, sealdomain.Request{
		PartnershipID: request.PartnershipID, Event: sealdomain.Event(request.Event),
		Document: sealdomain.Document{
			Type: request.DocumentType, ID: request.DocumentID, Number: request.DocumentNumber,
		},
		Content: request.Content, ActorUserID: request.ActorUserID, ActorOrgID: request.ActorOrgID, At: request.At,
	})
	if err != nil {
		return fmt.Errorf("seal %s: %w", request.DocumentType, err)
	}
	return nil
}

type agreementFounder struct {
	founder *agreementservice.Founder
}

func newAgreementFounder(pool *database.Pool, recorder *sealservice.Recorder) agreementFounder {
	return agreementFounder{founder: agreementservice.NewFounder(agreementservice.FounderDeps{
		Agreements: agreementrepository.NewAgreements(pool), Sealer: agreementSealer{recorder: recorder},
		IDs: idgen.UUIDv7{},
	})}
}

func (a agreementFounder) Found(ctx context.Context, founding proposaldomain.Founding) (string, error) {
	items := make([]agreementdomain.Item, 0, len(founding.Items))
	for _, item := range founding.Items {
		items = append(items, agreementdomain.Item{ProductID: item.ProductID, Terms: item.Terms})
	}
	agreement, err := a.founder.Found(ctx, agreementdomain.FoundingInput{
		ProposalID: founding.ProposalID, ProposalNumber: founding.ProposalNumber, ConsignorID: founding.ConsignorID,
		StoreID: founding.StoreID, Terms: founding.Terms, Items: items, ProposedBy: founding.ProposedBy,
		AcceptedBy: founding.AcceptedBy, AcceptorOrgID: founding.AcceptorOrgID, At: founding.At,
	})
	if errors.Is(err, agreementdomain.ErrAlreadyPartners) {
		return "", proposaldomain.ErrAlreadyPartners
	}
	if err != nil {
		return "", fmt.Errorf("found agreement: %w", err)
	}
	return agreement.PartnershipID, nil
}
