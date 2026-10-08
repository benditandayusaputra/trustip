package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
)

type directScene struct {
	scene
	announcer *fakeAnnouncer
}

func newDirectScene() directScene {
	world := newScene(sentAt())
	announcer := &fakeAnnouncer{}
	world.deps.Announcer = announcer
	world.direct.parties = domain.DirectParties{
		SenderName: "Roti Pak Anton",
		Target:     domain.Target{OrganizationID: storeID, Name: "Warung Bu Sri", Type: actor.Store},
	}
	return directScene{scene: world, announcer: announcer}
}

func propose(world directScene, sender actor.Actor) error {
	_, err := service.NewSendProposal(world.deps).Execute(t0(), sender, service.DirectCommand{
		TargetID: storeID, Input: invitationInput(),
	})
	if err != nil {
		return fmt.Errorf("propose: %w", err)
	}
	return nil
}

func TestSendProposal_FRPRP01_ConsignorSendsAProposalToAListedStore(t *testing.T) {
	world := newDirectScene()

	err := propose(world, owner(consignorID, actor.Consignor))

	opening := world.proposals.invitations[""]
	notices := world.announcer.notices
	if err != nil || opening.AwaitingID != storeID || len(world.direct.consumed) != 1 || len(notices) != 1 ||
		notices[0].Body != "Roti Pak Anton mengajukan kerja sama." || len(world.audit.entries) != 1 {
		t.Fatalf("opening = %+v, err = %v, notices = %+v", opening, err, world.announcer.notices)
	}
}

func TestSendProposal_BRAUTHZ01_StaffCannotPropose(t *testing.T) {
	world := newDirectScene()
	staff := owner(consignorID, actor.Consignor)
	staff.Membership.Role = actor.Staff

	err := propose(world, staff)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestSendProposal_BRAUTHZ02_UnlistedTargetIsNotFound(t *testing.T) {
	world := newDirectScene()
	world.direct.parties.Target.OrganizationID = "someone-else"

	err := propose(world, owner(consignorID, actor.Consignor))

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestSendProposal_BRPRP05_PartnersAreSentToAnAmendment(t *testing.T) {
	world := newDirectScene()
	world.direct.parties.ArePartners = true

	err := propose(world, owner(consignorID, actor.Consignor))

	if apperror.From(err).Code() != apperror.CodeAlreadyPartners || len(world.direct.consumed) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestSendProposal_BRPRP06_EleventhProposalOfTheDayIsRefused(t *testing.T) {
	world := newDirectScene()
	world.direct.exhausted = true

	err := propose(world, owner(consignorID, actor.Consignor))

	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeProposalLimitReached ||
		appErr.Message() != "Batas 10 ajuan baru per hari tercapai. Coba lagi besok." {
		t.Fatalf("err = %v", err)
	}
}

func TestSendProposal_BRPRP06_RecentlyDeclinedSenderWaitsSevenDays(t *testing.T) {
	world := newDirectScene()
	world.direct.parties.LastDeclinedAt = sentAt().Add(-2 * 24 * time.Hour)

	err := propose(world, owner(consignorID, actor.Consignor))

	if apperror.From(err).Code() != apperror.CodeProposalCooldown || len(world.direct.consumed) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestSendInvitation_BRPRP06_InvitationsUseTheSameDailyQuota(t *testing.T) {
	world := newScene(sentAt())
	world.direct.exhausted = true

	_, err := service.NewSendInvitation(world.deps).Execute(t0(), owner(consignorID, actor.Consignor),
		invitationInput())

	if apperror.From(err).Code() != apperror.CodeProposalLimitReached {
		t.Fatalf("err = %v", err)
	}
}

func t0() context.Context {
	return context.Background()
}
