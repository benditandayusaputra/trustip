package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	agreementdomain "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	agreementhandler "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/handler"
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

func (s sealSigner) Verify(digest []byte, signature sealdomain.Signature) (bool, error) {
	isValid, err := s.sealer.Verify(digest, security.Signature(signature))
	if err != nil {
		return false, fmt.Errorf("verify seal signature: %w", err)
	}
	return isValid, nil
}

func (s sealSigner) PublicKeys() []sealdomain.PublicKey {
	published := s.sealer.PublicKeys()
	keys := make([]sealdomain.PublicKey, 0, len(published))
	for _, key := range published {
		keys = append(keys, sealdomain.PublicKey(key))
	}
	return keys
}

type sealComponents struct {
	recorder *sealservice.Recorder
	verifier *sealservice.Verifier
	public   *sealservice.VerifyPublicSeal
	keys     *sealservice.ListPublicKeys
	chains   *sealservice.CheckChains
}

func buildSealComponents(pool *database.Pool, keys security.ActiveKeyring) sealComponents {
	seals, signer := sealrepository.NewSeals(pool), sealSigner{sealer: security.NewSealer(keys)}
	return sealComponents{
		recorder: sealservice.NewRecorder(seals, signer, idgen.UUIDv7{}),
		verifier: sealservice.NewVerifier(seals, signer),
		public:   sealservice.NewVerifyPublicSeal(seals, signer),
		keys:     sealservice.NewListPublicKeys(signer),
		chains:   sealservice.NewCheckChains(seals, signer),
	}
}

type sealEvent struct {
	Event         string
	ActorOrgName  string
	ActorUserName string
	At            time.Time
	Code          string
}

type sealCheck struct {
	DocumentType string
	DocumentID   string
	Content      map[string]any
}

type documentSealChecker struct {
	verifier *sealservice.Verifier
}

func (c documentSealChecker) check(ctx context.Context, check sealCheck) (bool, []sealEvent, error) {
	verdict, err := c.verifier.Check(ctx, sealdomain.Check{
		Document: sealdomain.Document{Type: check.DocumentType, ID: check.DocumentID}, Content: check.Content,
	})
	if err != nil {
		return false, nil, fmt.Errorf("check seal: %w", err)
	}
	events := make([]sealEvent, 0, len(verdict.Trail))
	for _, entry := range verdict.Trail {
		events = append(events, sealEvent{
			Event: string(entry.Seal.Entry.Event), ActorOrgName: entry.ActorOrgName, ActorUserName: entry.ActorUserName,
			At: entry.Seal.Entry.OccurredAt, Code: entry.Seal.Code,
		})
	}
	return verdict.IsValid, events, nil
}

type agreementSealChecker struct {
	documentSealChecker
}

func (c agreementSealChecker) Check(
	ctx context.Context, check agreementdomain.SealCheck,
) (agreementdomain.SealVerdict, error) {
	isValid, trail, err := c.check(ctx, sealCheck(check))
	if err != nil {
		return agreementdomain.SealVerdict{}, err
	}
	events := make([]agreementdomain.SealEvent, 0, len(trail))
	for _, event := range trail {
		events = append(events, agreementdomain.SealEvent(event))
	}
	return agreementdomain.SealVerdict{IsValid: isValid, Events: events}, nil
}

type agreementComponents struct {
	reader  agreementservice.AgreementReader
	checker agreementservice.SealChecker
}

func buildAgreementComponents(pool *database.Pool, seals sealComponents) agreementComponents {
	return agreementComponents{
		reader:  agreementrepository.NewAgreements(pool),
		checker: agreementSealChecker{documentSealChecker{verifier: seals.verifier}},
	}
}

func agreementDeps(components agreementComponents, routes proposalRoutes) agreementhandler.Deps {
	return agreementhandler.Deps{
		Authenticate: routes.authenticate,
		Get:          agreementservice.NewGetAgreement(components.reader, components.checker),
		History:      agreementservice.NewListVersions(components.reader),
		SecurityLog:  routes.securityLog,
	}
}

type sealRequest struct {
	PartnershipID  string
	Event          string
	DocumentType   string
	DocumentID     string
	DocumentNumber string
	Content        map[string]any
	ActorUserID    string
	ActorOrgID     string
	At             time.Time
}

type documentSealer struct {
	recorder *sealservice.Recorder
}

func (s documentSealer) seal(ctx context.Context, request sealRequest) error {
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

type agreementSealer struct {
	documentSealer
}

func (s agreementSealer) Seal(ctx context.Context, request agreementdomain.SealRequest) error {
	return s.seal(ctx, sealRequest(request))
}

type agreementKeeper struct {
	founder *agreementservice.Founder
}

func newAgreementKeeper(pool *database.Pool, seals sealComponents) agreementKeeper {
	return agreementKeeper{founder: agreementservice.NewFounder(agreementservice.FounderDeps{
		Agreements: agreementrepository.NewAgreements(pool),
		Sealer:     agreementSealer{documentSealer{recorder: seals.recorder}},
		IDs:        idgen.UUIDv7{},
	})}
}

func (a agreementKeeper) Found(ctx context.Context, founding proposaldomain.Founding) (string, error) {
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

func (a agreementKeeper) Amend(ctx context.Context, founding proposaldomain.Founding) (int, error) {
	changes := make([]agreementdomain.Change, 0, len(founding.Items))
	for _, item := range founding.Items {
		changes = append(changes, agreementdomain.Change{
			ProductID: item.ProductID, Action: agreementdomain.ChangeAction(item.Action), Terms: item.Terms,
		})
	}
	agreement, err := a.founder.Amend(ctx, agreementdomain.AmendmentInput{
		PartnershipID: founding.PartnershipID, ProposalID: founding.ProposalID,
		ProposalNumber: founding.ProposalNumber, Terms: founding.Terms, Changes: changes,
		ProposedBy: founding.ProposedBy, AcceptedBy: founding.AcceptedBy, AcceptorOrgID: founding.AcceptorOrgID,
		At: founding.At,
	})
	if err != nil {
		return 0, fmt.Errorf("amend agreement: %w", err)
	}
	return agreement.Version, nil
}

type currentAgreements struct {
	reader agreementservice.AgreementReader
}

func (c currentAgreements) CurrentAgreement(
	ctx context.Context, key proposaldomain.PartnershipKey,
) (proposaldomain.CurrentAgreement, error) {
	view, err := c.reader.GetAgreement(ctx, agreementdomain.VersionKey{
		OrganizationID: key.OrganizationID, PartnershipID: key.PartnershipID,
	})
	if err != nil {
		return proposaldomain.CurrentAgreement{}, fmt.Errorf("read current agreement: %w", err)
	}
	items := map[string]proposaldomain.CurrentItem{}
	for _, item := range view.Agreement.Items {
		items[item.ProductID] = proposaldomain.CurrentItem{IsActive: item.IsActive, Terms: item.Terms}
	}
	return proposaldomain.CurrentAgreement{
		PartnershipID: key.PartnershipID, IsActive: view.IsPartnershipActive, Terms: view.Agreement.Terms,
		Consignor: proposaldomain.Party{OrganizationID: view.Consignor.OrganizationID, Name: view.Consignor.Name},
		Store:     proposaldomain.Party{OrganizationID: view.Store.OrganizationID, Name: view.Store.Name},
		Items:     items,
	}, nil
}
