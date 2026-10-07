package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type declineRequest struct {
	Reason string `json:"reason"`
}

func decline(decliner Decliner) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		proposalID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read proposal id: %w", err)
		}
		var request declineRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read decline: %w", err)
		}
		reason := httpserver.MultilineInput("reason", request.Reason)
		if err := httpserver.RejectUncleanInputs(reason); err != nil {
			return fmt.Errorf("check decline reason: %w", err)
		}
		proposal, err := decliner.Execute(c.Context(), signedIn,
			service.DeclineCommand{ProposalID: proposalID, Reason: reason.Value})
		if err != nil {
			return fmt.Errorf("decline proposal: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toProposalResponse(proposal)})
	}
}

type summaryResponse struct {
	ID              string  `json:"id"`
	Number          string  `json:"number"`
	Kind            string  `json:"kind"`
	Status          string  `json:"status"`
	FromOrgID       string  `json:"from_org_id"`
	AwaitingOrgID   *string `json:"awaiting_org_id"`
	CounterpartName string  `json:"counterpart_name"`
	ExpiresAt       string  `json:"expires_at"`
	CreatedAt       string  `json:"created_at"`
}

func list(lister Lister) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		page, err := lister.Execute(c.Context(), signedIn,
			service.ListRequest{Box: c.Query("box"), Cursor: c.Query("cursor"), Limit: c.Query("limit")})
		if err != nil {
			return fmt.Errorf("list proposals: %w", err)
		}
		items := make([]summaryResponse, 0, len(page.Items))
		for _, summary := range page.Items {
			items = append(items, toSummaryResponse(summary))
		}
		return httpserver.RespondList(c, httpserver.Page[summaryResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
	}
}

func toSummaryResponse(summary domain.Summary) summaryResponse {
	return summaryResponse{
		ID: summary.ID, Number: summary.Number, Kind: string(summary.Kind), Status: string(summary.Status),
		FromOrgID: summary.FromOrgID, AwaitingOrgID: optional(summary.AwaitingOrgID),
		CounterpartName: summary.CounterpartName, ExpiresAt: summary.ExpiresAt.UTC().Format(time.RFC3339),
		CreatedAt: summary.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func amend(sender AmendmentSender) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		var request invitationRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read amendment: %w", err)
		}
		input, err := request.toInput()
		if err != nil {
			return err
		}
		proposal, err := sender.Execute(c.Context(), signedIn,
			service.AmendmentCommand{PartnershipID: partnershipID, Input: input})
		if err != nil {
			return fmt.Errorf("send amendment: %w", err)
		}
		c.Location(httpserver.APIPrefix + proposalsPath + "/" + proposal.ID)
		return httpserver.Respond(c, httpserver.Reply{Status: fiber.StatusCreated, Data: toProposalResponse(proposal)})
	}
}
