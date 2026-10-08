package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

func claimedInvitation(t *testing.T, scene world) domain.Invitation {
	t.Helper()
	created := invitation(1, consignorID, breadID)
	scene.invite(created)
	scene.must(scene.proposals.ClaimInvitation(t.Context(), domain.Claim{
		ProposalID: created.IDs.ProposalID, ConsignorID: consignorID, StoreID: storeID, ClaimerID: storeID, At: base(),
	}))
	return created
}

func counteroffer(created domain.Invitation, previousNo int) domain.Counteroffer {
	draft := created.Draft
	draft.Message = "Bisa lebih murah?"
	return domain.Counteroffer{
		ProposalID: created.IDs.ProposalID, RevisionID: "0192f0c4-0000-7000-9000-000000000099",
		RevisionNo: previousNo + 1, PreviousNo: previousNo, AuthorOrgID: storeID, AuthorUserID: userID,
		AwaitingOrgID: consignorID, ConsignorID: consignorID, Draft: draft, At: base().Add(time.Hour),
		ExpiresAt: base().Add(domain.OpenFor),
	}
}

func TestSaveCounter_FRPRP04_NewRevisionMovesTheTurn(t *testing.T) {
	scene := newWorld(t)
	created := claimedInvitation(t, scene)

	scene.must(scene.proposals.SaveCounter(t.Context(), counteroffer(created, 1)))

	key := domain.ProposalKey{OrganizationID: consignorID, ProposalID: created.IDs.ProposalID}
	proposal, err := scene.proposals.GetProposal(t.Context(), key)
	revisions, listErr := scene.proposals.ListRevisions(t.Context(), key)
	if err != nil || listErr != nil || proposal.Revision.No != 2 || proposal.AwaitingOrgID != consignorID ||
		len(revisions) != 2 || len(revisions[1].Items) != 1 || revisions[1].Message != "Bisa lebih murah?" {
		t.Fatalf("proposal = %+v, revisions = %+v, err = %v %v", proposal, revisions, err, listErr)
	}
}

func TestSaveCounter_StaleRevisionIsAlreadyDecided(t *testing.T) {
	scene := newWorld(t)
	created := claimedInvitation(t, scene)
	scene.must(scene.proposals.SaveCounter(t.Context(), counteroffer(created, 1)))

	err := scene.proposals.SaveCounter(t.Context(), counteroffer(created, 1))

	if !errors.Is(err, domain.ErrAlreadyDecided) {
		t.Fatalf("err = %v", err)
	}
}

func TestListRevisions_BRAUTHZ02_OtherOrganizationsSeeNothing(t *testing.T) {
	scene := newWorld(t)
	created := claimedInvitation(t, scene)

	_, err := scene.proposals.ListRevisions(t.Context(), domain.ProposalKey{
		OrganizationID: otherStore, ProposalID: created.IDs.ProposalID,
	})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}
