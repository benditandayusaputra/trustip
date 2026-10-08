package service_test

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
)

const proposalID = "0192f0c4-0000-7000-8000-000000000301"

type decidingProposals struct {
	*fakeProposals
	stored    domain.Proposal
	decisions []domain.Decision
	failSave  error
	summaries []domain.Summary
	listed    []domain.ListQuery
}

func (d *decidingProposals) LockProposal(_ context.Context, key domain.ProposalKey) (domain.Proposal, error) {
	if key.OrganizationID != consignorID && key.OrganizationID != storeID {
		return domain.Proposal{}, apperror.NotFound()
	}
	return d.stored, nil
}

func (d *decidingProposals) GetProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error) {
	proposal, err := d.LockProposal(ctx, key)
	if len(d.decisions) > 0 {
		proposal.Status = d.decisions[len(d.decisions)-1].Status
	}
	return proposal, err
}

func (d *decidingProposals) SaveDecision(_ context.Context, decision domain.Decision) error {
	if d.failSave != nil {
		return d.failSave
	}
	d.decisions = append(d.decisions, decision)
	return nil
}

func (d *decidingProposals) ListProposals(_ context.Context, query domain.ListQuery) ([]domain.Summary, error) {
	d.listed = append(d.listed, query)
	return d.summaries, nil
}

type fakeFounder struct {
	foundings  []domain.Founding
	amendments []domain.Founding
	fail       error
}

func (f *fakeFounder) Amend(_ context.Context, founding domain.Founding) (int, error) {
	f.amendments = append(f.amendments, founding)
	return 2, nil
}

func (f *fakeFounder) Found(_ context.Context, founding domain.Founding) (string, error) {
	if f.fail != nil {
		return "", f.fail
	}
	f.foundings = append(f.foundings, founding)
	return "0192f0c4-0000-7000-8000-000000000501", nil
}

type fakeAnnouncer struct{ notices []domain.Notice }

func (f *fakeAnnouncer) Announce(_ context.Context, notice domain.Notice) error {
	f.notices = append(f.notices, notice)
	return nil
}

type decisionScene struct {
	proposals *decidingProposals
	founder   *fakeFounder
	announcer *fakeAnnouncer
	audit     *fakeAudit
	deps      service.ProposalDeps
}

func newDecisionScene() decisionScene {
	proposals := &decidingProposals{fakeProposals: newFakeProposals(), stored: domain.Proposal{
		ID: proposalID, Number: "AJ-202610-0001", Kind: domain.New, Status: domain.Open,
		Consignor: domain.Party{OrganizationID: consignorID, Name: "Roti Pak Anton"},
		Store:     domain.Party{OrganizationID: storeID, Name: "Warung Bu Sri"},
		FromOrgID: consignorID, AwaitingOrgID: storeID, ExpiresAt: sentAt().Add(domain.OpenFor),
		Revision: domain.Revision{
			No: 1, AuthorOrgID: consignorID, AuthorUserID: "u-consignor",
			Terms: terms.General{Schedule: terms.ScheduleWeekly, Rhythm: terms.RhythmDaily},
		},
	}}
	founder, announcer, trail := &fakeFounder{}, &fakeAnnouncer{}, &fakeAudit{}
	return decisionScene{
		proposals: proposals, founder: founder, announcer: announcer, audit: trail,
		deps: service.ProposalDeps{
			Proposals: proposals, Audit: trail, Tx: inlineTx{}, Clock: fixedClock{at: sentAt()}, IDs: &sequenceIDs{},
			Codes: fixedCode{}, Keeper: founder, Announcer: announcer, Current: currentReader{}, Direct: &fakeDirect{},
		},
	}
}

func TestAcceptProposal_BRPRP03_FoundsThePartnershipAndTellsTheAuthor(t *testing.T) {
	scene := newDecisionScene()

	proposal, err := service.NewAcceptProposal(scene.deps).Execute(t.Context(), owner(storeID, actor.Store), proposalID)

	decision := scene.proposals.decisions[0]
	if err != nil || proposal.Status != domain.Accepted || len(scene.founder.foundings) != 1 ||
		decision.PartnershipID != "0192f0c4-0000-7000-8000-000000000501" || len(scene.announcer.notices) != 1 ||
		scene.announcer.notices[0].OrganizationID != consignorID ||
		scene.audit.entries[0].Action != audit.ProposalAccepted {
		t.Fatalf("proposal = %+v, decision = %+v, err = %v", proposal, decision, err)
	}
}

func TestAcceptProposal_BRAUTHZ01_StaffCannotDecide(t *testing.T) {
	scene := newDecisionScene()
	staff := owner(storeID, actor.Store)
	staff.Membership.Role = actor.Staff

	_, err := service.NewAcceptProposal(scene.deps).Execute(t.Context(), staff, proposalID)

	if apperror.From(err).Code() != apperror.CodeForbidden || len(scene.founder.foundings) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestAcceptProposal_BRPRP02_SenderCannotAcceptTheirOwnRevision(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewAcceptProposal(scene.deps).Execute(t.Context(), owner(consignorID, actor.Consignor),
		proposalID)

	if apperror.From(err).Code() != apperror.CodeNotYourTurn || len(scene.founder.foundings) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestAcceptProposal_BRPRT01_PartnersAlreadyActiveAreRefused(t *testing.T) {
	scene := newDecisionScene()
	scene.founder.fail = domain.ErrAlreadyPartners

	_, err := service.NewAcceptProposal(scene.deps).Execute(t.Context(), owner(storeID, actor.Store), proposalID)

	if apperror.From(err).Code() != apperror.CodeAlreadyPartners {
		t.Fatalf("err = %v", err)
	}
}

func TestAcceptProposal_ConcurrentDecisionLoses(t *testing.T) {
	scene := newDecisionScene()
	scene.proposals.failSave = domain.ErrAlreadyDecided

	_, err := service.NewAcceptProposal(scene.deps).Execute(t.Context(), owner(storeID, actor.Store), proposalID)

	if apperror.From(err).Code() != apperror.CodeProposalDecided {
		t.Fatalf("err = %v", err)
	}
}

func TestDeclineProposal_FRPRP03_StoresTheReasonAndTellsTheAuthor(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewDeclineProposal(scene.deps).Execute(t.Context(), owner(storeID, actor.Store),
		service.DeclineCommand{ProposalID: proposalID, Reason: "Rak penuh."})

	if err != nil || scene.proposals.decisions[0].Reason != "Rak penuh." ||
		scene.announcer.notices[0].Type != "PROPOSAL_DECLINED" || len(scene.founder.foundings) != 0 {
		t.Fatalf("decisions = %+v, err = %v", scene.proposals.decisions, err)
	}
}

func TestWithdrawProposal_FRPRP05_SenderWithdrawsWithoutNotifying(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewWithdrawProposal(scene.deps).Execute(t.Context(), owner(consignorID, actor.Consignor),
		proposalID)

	if err != nil || scene.proposals.decisions[0].Status != domain.Withdrawn || len(scene.announcer.notices) != 0 {
		t.Fatalf("decisions = %+v, err = %v", scene.proposals.decisions, err)
	}
}

func TestWithdrawProposal_BRAUTHZ02_StrangerGetsNotFound(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewWithdrawProposal(scene.deps).Execute(t.Context(), owner("elsewhere", actor.Consignor),
		proposalID)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestListProposals_FRPRP06_OutgoingBoxForTheSignedInOrganization(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewListProposals(scene.proposals).Execute(t.Context(), owner(consignorID, actor.Consignor),
		service.ListRequest{Box: "OUTGOING", Limit: "5"})

	query := scene.proposals.listed[0]
	if err != nil || query.OrganizationID != consignorID || query.Box != domain.Outgoing || query.Limit != 6 {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
}

func TestListProposals_UnknownBoxIsAValidationFailure(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewListProposals(scene.proposals).Execute(t.Context(), owner(consignorID, actor.Consignor),
		service.ListRequest{Box: "SPAM"})

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

type currentReader struct{ isEnded bool }

func (c currentReader) CurrentAgreement(_ context.Context, key domain.PartnershipKey) (domain.CurrentAgreement, error) {
	if key.OrganizationID != consignorID && key.OrganizationID != storeID {
		return domain.CurrentAgreement{}, apperror.NotFound()
	}
	return domain.CurrentAgreement{
		PartnershipID: key.PartnershipID, IsActive: !c.isEnded,
		Consignor: domain.Party{OrganizationID: consignorID, Name: "Roti Pak Anton"},
		Store:     domain.Party{OrganizationID: storeID, Name: "Warung Bu Sri"},
		Terms:     terms.General{Schedule: terms.ScheduleWeekly, Rhythm: terms.RhythmDaily},
		Items:     map[string]domain.CurrentItem{breadID: {IsActive: true}},
	}, nil
}

func amendmentCommand() service.AmendmentCommand {
	return service.AmendmentCommand{PartnershipID: "0192f0c4-0000-7000-8000-000000000501", Input: domain.RevisionInput{
		Terms: terms.Input{SettlementSchedule: "WEEKLY", CountRhythm: "DAILY"},
		Items: []domain.ItemInput{{ProductID: breadID, Action: "STOP"}},
	}}
}

func TestSendAmendment_FRPRP08_StopsAProductAndTellsTheOtherOwner(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewSendAmendment(scene.deps).Execute(t.Context(), owner(consignorID, actor.Consignor),
		amendmentCommand())

	notice := scene.announcer.notices[0]
	if err != nil || notice.OrganizationID != storeID || notice.Type != "PROPOSAL_RECEIVED" ||
		notice.Body != "Roti Pak Anton mengajukan perubahan Kesepakatan (AJ-202610-0001)." {
		t.Fatalf("notice = %+v, err = %v", notice, err)
	}
}

func TestSendAmendment_BRAUTHZ01_StaffCannotAmend(t *testing.T) {
	scene := newDecisionScene()
	staff := owner(consignorID, actor.Consignor)
	staff.Membership.Role = actor.Staff

	_, err := service.NewSendAmendment(scene.deps).Execute(t.Context(), staff, amendmentCommand())

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestSendAmendment_BRAUTHZ02_StrangerGetsNotFound(t *testing.T) {
	scene := newDecisionScene()

	_, err := service.NewSendAmendment(scene.deps).Execute(t.Context(), owner("elsewhere", actor.Consignor),
		amendmentCommand())

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestAcceptProposal_BRPRP03_AmendmentMakesTheNextVersionAndTellsBothOwners(t *testing.T) {
	scene := newDecisionScene()
	scene.proposals.stored.Kind = domain.Amendment
	scene.proposals.stored.PartnershipID = "0192f0c4-0000-7000-8000-000000000501"

	_, err := service.NewAcceptProposal(scene.deps).Execute(t.Context(), owner(storeID, actor.Store), proposalID)

	if err != nil || len(scene.founder.amendments) != 1 || len(scene.founder.foundings) != 0 ||
		scene.proposals.decisions[0].AgreementVersion != 2 || len(scene.announcer.notices) != 3 {
		t.Fatalf("notices = %+v, err = %v", scene.announcer.notices, err)
	}
}
