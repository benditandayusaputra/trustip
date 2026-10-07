package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	invitationsPath = "/invitations"
	invitationPath  = invitationsPath + "/:code"
	claimPath       = invitationPath + "/claim"
	proposalsPath   = "/proposals"
	proposalPath    = proposalsPath + "/:id"
	acceptPath      = proposalPath + "/accept"
	declinePath     = proposalPath + "/decline"
	withdrawPath    = proposalPath + "/withdraw"
)

type InvitationSender interface {
	Execute(ctx context.Context, signedIn actor.Actor, input domain.RevisionInput) (domain.Proposal, error)
}

type InvitationViewer interface {
	Execute(ctx context.Context, signedIn actor.Actor, code string) (domain.InvitationView, error)
}

type ProposalByKey interface {
	Execute(ctx context.Context, signedIn actor.Actor, key string) (domain.Proposal, error)
}

type Decliner interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.DeclineCommand) (domain.Proposal, error)
}

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest) (service.Page, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Send         InvitationSender
	View         InvitationViewer
	Claim        ProposalByKey
	Get          ProposalByKey
	Accept       ProposalByKey
	Decline      Decliner
	Withdraw     ProposalByKey
	List         Lister
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(invitationsPath, deps.Authenticate, writeLimit, send(deps.Send))
	api.Get(invitationPath, deps.Authenticate, view(deps.View))
	api.Post(claimPath, deps.Authenticate, writeLimit, byKey(deps.Claim, "code"))
	api.Get(proposalPath, deps.Authenticate, byID(deps.Get))
	api.Get(proposalsPath, deps.Authenticate, list(deps.List))
	api.Post(acceptPath, deps.Authenticate, writeLimit, byID(deps.Accept))
	api.Post(declinePath, deps.Authenticate, writeLimit, decline(deps.Decline))
	api.Post(withdrawPath, deps.Authenticate, writeLimit, byID(deps.Withdraw))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func send(sender InvitationSender) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		var request invitationRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read invitation: %w", err)
		}
		input, err := request.toInput()
		if err != nil {
			return err
		}
		proposal, err := sender.Execute(c.Context(), signedIn, input)
		if err != nil {
			return fmt.Errorf("send invitation: %w", err)
		}
		c.Location(httpserver.APIPrefix + proposalsPath + "/" + proposal.ID)
		return httpserver.Respond(c, httpserver.Reply{Status: fiber.StatusCreated, Data: toProposalResponse(proposal)})
	}
}

func view(viewer InvitationViewer) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		invitation, err := viewer.Execute(c.Context(), signedIn, c.Params("code"))
		if err != nil {
			return fmt.Errorf("view invitation: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: invitationResponse{
			SenderName: invitation.Sender.Name, SenderType: string(invitation.Sender.Type),
			ItemCount: invitation.ItemCount, ExpiresAt: invitation.ExpiresAt.UTC().Format(time.RFC3339),
		}})
	}
}

func byKey(reader ProposalByKey, param string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		proposal, err := reader.Execute(c.Context(), signedIn, c.Params(param))
		if err != nil {
			return fmt.Errorf("read proposal by %s: %w", param, err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toProposalResponse(proposal)})
	}
}

func byID(reader ProposalByKey) fiber.Handler {
	byParam := byKey(reader, "id")
	return func(c fiber.Ctx) error {
		if _, err := httpserver.IDParam(c, "id"); err != nil {
			return fmt.Errorf("read proposal id: %w", err)
		}
		return byParam(c)
	}
}
