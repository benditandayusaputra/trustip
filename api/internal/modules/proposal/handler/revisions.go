package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type Counterer interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.CounterCommand) (domain.Proposal, error)
}

type RevisionLister interface {
	Execute(ctx context.Context, signedIn actor.Actor, proposalID string) ([]domain.RevisionRecord, error)
}

type changedItemResponse struct {
	itemResponse
	Change string `json:"change"`
}

type revisionRecordResponse struct {
	RevisionNo  int                   `json:"revision_no"`
	AuthorOrgID string                `json:"author_org_id"`
	AuthorName  string                `json:"author_name"`
	Terms       termsBody             `json:"terms"`
	TermChanges []string              `json:"term_changes"`
	Message     string                `json:"message"`
	Items       []changedItemResponse `json:"items"`
	CreatedAt   string                `json:"created_at"`
}

func counter(counterer Counterer) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		proposalID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read proposal id: %w", err)
		}
		var request invitationRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read counter: %w", err)
		}
		input, err := request.toInput()
		if err != nil {
			return err
		}
		proposal, err := counterer.Execute(c.Context(), signedIn, service.CounterCommand{
			ProposalID: proposalID, Input: input,
		})
		if err != nil {
			return fmt.Errorf("counter proposal: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toProposalResponse(proposal)})
	}
}

func revisions(lister RevisionLister) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		proposalID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read proposal id: %w", err)
		}
		records, err := lister.Execute(c.Context(), signedIn, proposalID)
		if err != nil {
			return fmt.Errorf("list revisions: %w", err)
		}
		items := make([]revisionRecordResponse, 0, len(records))
		for _, record := range records {
			items = append(items, toRevisionRecordResponse(record))
		}
		return httpserver.RespondList(c, httpserver.Page[revisionRecordResponse]{Items: items})
	}
}

func toRevisionRecordResponse(record domain.RevisionRecord) revisionRecordResponse {
	revision := toRevisionResponse(record.Revision)
	items := make([]changedItemResponse, 0, len(record.Changes.Items))
	for _, difference := range record.Changes.Items {
		items = append(items, changedItemResponse{
			itemResponse: toItemResponse(difference.Item), Change: string(difference.Change),
		})
	}
	return revisionRecordResponse{
		RevisionNo: revision.RevisionNo, AuthorOrgID: revision.AuthorOrgID, AuthorName: record.AuthorName,
		Terms: revision.Terms, TermChanges: append([]string{}, record.Changes.TermFields...),
		Message: revision.Message, Items: items, CreatedAt: record.Revision.At.UTC().Format(time.RFC3339),
	}
}
