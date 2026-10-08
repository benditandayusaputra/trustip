package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
)

type fakeBargains struct {
	offers    []domain.Counteroffer
	revisions []domain.Revision
	raced     bool
}

func (f *fakeBargains) SaveCounter(_ context.Context, offer domain.Counteroffer) error {
	if f.raced {
		return domain.ErrAlreadyDecided
	}
	f.offers = append(f.offers, offer)
	return nil
}

func (f *fakeBargains) ListRevisions(_ context.Context, key domain.ProposalKey) ([]domain.Revision, error) {
	if key.OrganizationID != consignorID && key.OrganizationID != storeID {
		return nil, apperror.NotFound()
	}
	return f.revisions, nil
}

func counterScene() (decisionScene, *fakeBargains) {
	scene := newDecisionScene()
	bargains := &fakeBargains{}
	scene.deps.Bargains = bargains
	return scene, bargains
}

func counterAs(scene decisionScene, counterer actor.Actor) error {
	_, err := service.NewCounterProposal(scene.deps).Execute(context.Background(), counterer, service.CounterCommand{
		ProposalID: proposalID, Input: invitationInput(),
	})
	if err != nil {
		return fmt.Errorf("counter: %w", err)
	}
	return nil
}

func TestCounterProposal_FRPRP04_StoreCountersAndTheConsignorIsTold(t *testing.T) {
	scene, bargains := counterScene()

	err := counterAs(scene, owner(storeID, actor.Store))

	if err != nil || len(bargains.offers) != 1 || bargains.offers[0].AwaitingOrgID != consignorID ||
		len(scene.announcer.notices) != 1 || scene.announcer.notices[0].Type != "PROPOSAL_COUNTERED" ||
		scene.audit.entries[0].Action != audit.ProposalCountered {
		t.Fatalf("offers = %+v, notices = %+v, err = %v", bargains.offers, scene.announcer.notices, err)
	}
}

func TestCounterProposal_BRPRP02_SenderOfTheLastRevisionCannotCounterItself(t *testing.T) {
	scene, bargains := counterScene()

	err := counterAs(scene, owner(consignorID, actor.Consignor))

	if apperror.From(err).Code() != apperror.CodeNotYourTurn || len(bargains.offers) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestCounterProposal_ConcurrentCounterIsAlreadyDecided(t *testing.T) {
	scene, bargains := counterScene()
	bargains.raced = true

	err := counterAs(scene, owner(storeID, actor.Store))

	if apperror.From(err).Code() != apperror.CodeProposalDecided || len(scene.announcer.notices) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestCounterProposal_BRAUTHZ01_StaffCannotCounter(t *testing.T) {
	scene, _ := counterScene()
	staff := owner(storeID, actor.Store)
	staff.Membership.Role = actor.Staff

	if apperror.From(counterAs(scene, staff)).Code() != apperror.CodeForbidden {
		t.Fatal("staff countered")
	}
}

func TestListRevisions_FRPRP04_HistoryNamesTheAuthorOfEachRevision(t *testing.T) {
	scene, bargains := counterScene()
	bargains.revisions = []domain.Revision{scene.proposals.stored.Revision}

	history, err := service.NewListRevisions(scene.deps).Execute(context.Background(), owner(storeID, actor.Store),
		proposalID)

	if err != nil || len(history) != 1 || history[0].AuthorName != "Roti Pak Anton" {
		t.Fatalf("history = %+v, err = %v", history, err)
	}
}
