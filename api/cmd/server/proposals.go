package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	proposaldomain "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	proposalhandler "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/handler"
	proposalrepository "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/repository"
	proposalservice "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type proposalComponents struct {
	deps proposalservice.ProposalDeps
}

func inviteCodes() idgen.Codes {
	return idgen.NewCodes(proposaldomain.InviteAlphabet(), proposaldomain.InviteLength)
}

type proposalCollaborators struct {
	founder   proposalservice.AgreementFounder
	announcer proposalservice.Announcer
}

func buildProposalComponents(
	pool *database.Pool, clock proposalservice.Clock, collaborators proposalCollaborators,
) proposalComponents {
	return proposalComponents{deps: proposalservice.ProposalDeps{
		Founder:   collaborators.founder,
		Announcer: collaborators.announcer,
		Proposals: proposalrepository.NewProposals(pool),
		Audit:     newAuditTrail(auditrepository.NewRecords(pool)),
		Tx:        pool,
		Clock:     clock,
		IDs:       idgen.UUIDv7{},
		Codes:     inviteCodes(),
	}}
}

type proposalRoutes struct {
	authenticate fiber.Handler
	writeLimiter httpserver.RateLimiter
	securityLog  *slog.Logger
}

func proposalDeps(components proposalComponents, routes proposalRoutes) proposalhandler.Deps {
	deps := components.deps
	return proposalhandler.Deps{
		Authenticate: routes.authenticate,
		Send:         proposalservice.NewSendInvitation(deps),
		View:         proposalservice.NewViewInvitation(deps),
		Claim:        proposalservice.NewClaimInvitation(deps),
		Get:          proposalservice.NewGetProposal(deps.Proposals),
		Accept:       proposalservice.NewAcceptProposal(deps),
		Decline:      proposalservice.NewDeclineProposal(deps),
		Withdraw:     proposalservice.NewWithdrawProposal(deps),
		List:         proposalservice.NewListProposals(deps.Proposals),
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
