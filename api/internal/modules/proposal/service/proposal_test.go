package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
)

const (
	consignorID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	storeID     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	breadID     = "0192f0c4-0000-7000-8000-000000000101"
	inviteCode  = "ABCDEFGH23"
)

type fakeProposals struct {
	invitations map[string]domain.Opening
	claims      []domain.Claim
	sequences   []domain.SequenceKey
	partners    bool
	failCreate  error
	failClaim   error
}

func newFakeProposals() *fakeProposals {
	return &fakeProposals{invitations: map[string]domain.Opening{}}
}

func (f *fakeProposals) NextNumber(_ context.Context, key domain.SequenceKey) (int, error) {
	f.sequences = append(f.sequences, key)
	return len(f.sequences), nil
}

func (f *fakeProposals) Open(_ context.Context, opening domain.Opening) error {
	if f.failCreate != nil {
		return f.failCreate
	}
	f.invitations[opening.IDs.Code] = opening
	return nil
}

func (f *fakeProposals) ClaimInvitation(_ context.Context, claim domain.Claim) error {
	if f.failClaim != nil {
		return f.failClaim
	}
	f.claims = append(f.claims, claim)
	return nil
}

func (f *fakeProposals) GetInvitation(_ context.Context, code string) (domain.InvitationView, error) {
	invitation, ok := f.invitations[code]
	if !ok {
		return domain.InvitationView{}, apperror.NotFound()
	}
	view := domain.InvitationView{
		ProposalID: invitation.IDs.ProposalID, Status: domain.Open, ExpiresAt: invitation.ExpiresAt,
		Sender: domain.SenderView{
			OrganizationID: invitation.Sender.OrganizationID, Name: "Roti Pak Anton",
			Type: invitation.Sender.OrganizationType,
		},
		ConsignorID: invitation.ConsignorID, StoreID: invitation.StoreID, ItemCount: len(invitation.Draft.Items),
	}
	for _, claim := range f.claims {
		view.StoreID = claim.StoreID
	}
	return view, nil
}

func (f *fakeProposals) GetProposal(_ context.Context, key domain.ProposalKey) (domain.Proposal, error) {
	for _, invitation := range f.invitations {
		isParty := key.OrganizationID == invitation.Sender.OrganizationID || len(f.claims) > 0
		if invitation.IDs.ProposalID == key.ProposalID && isParty {
			return domain.Proposal{ID: key.ProposalID, Number: invitation.Number, InviteCode: invitation.IDs.Code}, nil
		}
	}
	return domain.Proposal{}, apperror.NotFound()
}

func (f *fakeProposals) LockProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error) {
	return f.GetProposal(ctx, key)
}

func (f *fakeProposals) SaveDecision(context.Context, domain.Decision) error { return nil }

func (f *fakeProposals) ListProposals(context.Context, domain.ListQuery) ([]domain.Summary, error) {
	return []domain.Summary{}, nil
}

func (f *fakeProposals) HasActivePartnership(context.Context, domain.Pair) (bool, error) {
	return f.partners, nil
}

type fakeDirect struct {
	exhausted bool
	parties   domain.DirectParties
	consumed  []domain.QuotaKey
}

func (f *fakeDirect) ConsumeQuota(_ context.Context, key domain.QuotaKey) error {
	if f.exhausted {
		return domain.ErrQuotaExhausted
	}
	f.consumed = append(f.consumed, key)
	return nil
}

func (f *fakeDirect) GetDirectParties(_ context.Context, pair domain.DirectPair) (domain.DirectParties, error) {
	if f.parties.Target.OrganizationID != pair.TargetID {
		return domain.DirectParties{}, apperror.NotFound()
	}
	return f.parties, nil
}

type fakeAudit struct{ entries []audit.Entry }

func (f *fakeAudit) Record(_ context.Context, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

func sentAt() time.Time { return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC) }

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

type fixedCode struct{}

func (fixedCode) NewCode() (string, error) { return inviteCode, nil }

type scene struct {
	proposals *fakeProposals
	direct    *fakeDirect
	audit     *fakeAudit
	deps      service.ProposalDeps
}

func newScene(at time.Time) scene {
	proposals, direct, trail := newFakeProposals(), &fakeDirect{}, &fakeAudit{}
	return scene{proposals: proposals, direct: direct, audit: trail, deps: service.ProposalDeps{
		Proposals: proposals, Direct: direct, Audit: trail, Tx: inlineTx{}, Clock: fixedClock{at: at},
		IDs: &sequenceIDs{}, Codes: fixedCode{},
	}}
}

func owner(organizationID string, organizationType actor.OrganizationType) actor.Actor {
	return actor.Actor{UserID: "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e", Membership: actor.Membership{
		OrganizationID: organizationID, OrganizationType: organizationType, Role: actor.Owner,
	}}
}

func invitationInput() domain.RevisionInput {
	return domain.RevisionInput{
		Terms: terms.Input{SettlementSchedule: "ON_COUNT", PaymentTermsDays: 3, CountRhythm: "DAILY"},
		Items: []domain.ItemInput{{ProductID: breadID, Terms: pricing.TermsInput{
			Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 1750,
		}}},
	}
}

func sendInvitation(t *testing.T, world scene) domain.Proposal {
	t.Helper()
	proposal, err := service.NewSendInvitation(world.deps).Execute(t.Context(), owner(consignorID, actor.Consignor),
		invitationInput())
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	return proposal
}

func TestSendInvitation_BRAUTHZ01_StaffCannotSendProposals(t *testing.T) {
	world := newScene(sentAt())
	staff := owner(consignorID, actor.Consignor)
	staff.Membership.Role = actor.Staff

	_, err := service.NewSendInvitation(world.deps).Execute(t.Context(), staff, invitationInput())

	if apperror.From(err).Code() != apperror.CodeForbidden || len(world.proposals.invitations) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestSendInvitation_BRPRP01_ProductOutsideTheCatalogIsAFieldError(t *testing.T) {
	world := newScene(sentAt())
	world.proposals.failCreate = domain.ErrUnknownProduct

	_, err := service.NewSendInvitation(world.deps).Execute(t.Context(), owner(consignorID, actor.Consignor),
		invitationInput())

	if fields := apperror.From(err).Fields(); len(fields) != 1 || fields[0].Field != "items" {
		t.Fatalf("err = %v", err)
	}
}

func TestSendInvitation_FRPRP07_NumbersPerSenderAndMonthAndAudits(t *testing.T) {
	world := newScene(sentAt())

	proposal := sendInvitation(t, world)

	want := domain.SequenceKey{OrganizationID: consignorID, Prefix: "AJ", Period: "202610"}
	if proposal.Number != "AJ-202610-0001" || proposal.InviteCode != inviteCode ||
		world.proposals.sequences[0] != want ||
		world.audit.entries[0].Action != audit.ProposalSent {
		t.Fatalf("proposal = %+v, sequences = %+v", proposal, world.proposals.sequences)
	}
}

func TestViewInvitation_BRPRP07_WrongTypeIsToldWhoItIsFor(t *testing.T) {
	world := newScene(sentAt())
	sendInvitation(t, world)

	_, err := service.NewViewInvitation(world.deps).Execute(t.Context(), owner("other", actor.Consignor), inviteCode)

	if apperror.From(err).Code() != apperror.CodeInviteWrongOrgType {
		t.Fatalf("err = %v", err)
	}
}

func TestViewInvitation_MalformedCodeIsNotFound(t *testing.T) {
	world := newScene(sentAt())

	_, err := service.NewViewInvitation(world.deps).Execute(t.Context(), owner(storeID, actor.Store), "abc")

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestClaimInvitation_BRPRP07_StoreBecomesTheOtherParty(t *testing.T) {
	world := newScene(sentAt())
	sent := sendInvitation(t, world)

	claimer := owner(storeID, actor.Store)

	proposal, err := service.NewClaimInvitation(world.deps).Execute(t.Context(), claimer, inviteCode)

	if err != nil || proposal.ID != sent.ID || len(world.proposals.claims) != 1 ||
		world.proposals.claims[0].StoreID != storeID ||
		world.audit.entries[1].Action != audit.InvitationClaimed {
		t.Fatalf("proposal = %+v, claims = %+v, err = %v", proposal, world.proposals.claims, err)
	}
}

func TestClaimInvitation_BRPRP05_ExistingPartnersAreSentToAnAmendment(t *testing.T) {
	world := newScene(sentAt())
	sendInvitation(t, world)
	world.proposals.partners = true

	_, err := service.NewClaimInvitation(world.deps).Execute(t.Context(), owner(storeID, actor.Store), inviteCode)

	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeAlreadyPartners || len(world.proposals.claims) != 0 ||
		appErr.Message() != "Anda sudah bermitra dengan Roti Pak Anton. Ajukan perubahan Kesepakatan." {
		t.Fatalf("err = %v", err)
	}
}

func TestClaimInvitation_BRPRP04_OpenProposalForThePairIsAConflict(t *testing.T) {
	world := newScene(sentAt())
	sendInvitation(t, world)
	world.proposals.failClaim = domain.ErrAlreadyOpen

	_, err := service.NewClaimInvitation(world.deps).Execute(t.Context(), owner(storeID, actor.Store), inviteCode)

	if apperror.From(err).Code() != apperror.CodeProposalAlreadyOpen {
		t.Fatalf("err = %v", err)
	}
}

func TestClaimInvitation_ConcurrentClaimLosesWithNotFound(t *testing.T) {
	world := newScene(sentAt())
	sendInvitation(t, world)
	world.proposals.failClaim = domain.ErrClaimRaced

	_, err := service.NewClaimInvitation(world.deps).Execute(t.Context(), owner(storeID, actor.Store), inviteCode)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestClaimInvitation_BRPRP08_ExpiredAfterFourteenDays(t *testing.T) {
	world := newScene(sentAt())
	sendInvitation(t, world)
	late := newScene(sentAt().Add(domain.OpenFor))
	late.proposals = world.proposals
	late.deps.Proposals = world.proposals

	_, err := service.NewClaimInvitation(late.deps).Execute(t.Context(), owner(storeID, actor.Store), inviteCode)

	if apperror.From(err).Code() != apperror.CodeProposalExpired {
		t.Fatalf("err = %v", err)
	}
}

func TestGetProposal_BRAUTHZ02_NonPartyGetsNotFound(t *testing.T) {
	world := newScene(sentAt())
	sent := sendInvitation(t, world)

	_, err := service.NewGetProposal(world.proposals).Execute(t.Context(), owner(storeID, actor.Store), sent.ID)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}
