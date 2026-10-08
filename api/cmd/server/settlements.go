package main

import (
	"context"
	"fmt"
	"time"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	countdomain "github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	countrepository "github.com/benditandayusaputra/trustip/api/internal/modules/count/repository"
	settlementdomain "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	settlementhandler "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/handler"
	settlementrepository "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/repository"
	settlementservice "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type settlementSealer struct {
	documentSealer
}

func (s settlementSealer) Seal(ctx context.Context, request settlementdomain.SealRequest) error {
	return s.seal(ctx, sealRequest(request))
}

type settlementSealChecker struct {
	documentSealChecker
}

func (c settlementSealChecker) Check(
	ctx context.Context, check settlementdomain.SealCheck,
) (settlementdomain.SealVerdict, error) {
	isValid, trail, err := c.check(ctx, sealCheck(check))
	if err != nil {
		return settlementdomain.SealVerdict{}, err
	}
	events := make([]settlementdomain.SealEvent, 0, len(trail))
	for _, event := range trail {
		events = append(events, settlementdomain.SealEvent(event))
	}
	return settlementdomain.SealVerdict{IsValid: isValid, Events: events}, nil
}

type settlementAnnouncer struct {
	organizationAnnouncer
}

func (a settlementAnnouncer) Announce(ctx context.Context, notice settlementdomain.Notice) error {
	return a.announce(ctx, organizationNotice(notice))
}

func buildSettlementDeps(pool *database.Pool, clock serverClock, seals sealComponents) settlementservice.Deps {
	return settlementservice.Deps{
		Settlements: settlementrepository.NewSettlements(pool),
		Sealer:      settlementSealer{documentSealer{recorder: seals.recorder}},
		Seals:       settlementSealChecker{documentSealChecker{verifier: seals.verifier}},
		Announcer:   settlementAnnouncer{newOrganizationAnnouncer(pool, clock)},
		Audit:       newAuditTrail(auditrepository.NewRecords(pool)),
		Approver:    countApprover{counts: countrepository.NewCounts(pool)},
		Tx:          pool,
		Clock:       clock,
		IDs:         idgen.UUIDv7{},
	}
}

func settlementDeps(deps settlementservice.Deps, routes tradeRoutes) settlementhandler.Deps {
	return settlementhandler.Deps{
		Authenticate: routes.authenticate,
		Get:          settlementservice.NewGetSettlement(deps),
		List:         settlementservice.NewListSettlements(deps),
		Running:      settlementservice.NewGetRunning(deps),
		Payments: settlementhandler.PaymentDeps{
			Record:       settlementservice.NewRecordPayment(deps),
			Confirm:      settlementservice.NewConfirmPayment(deps),
			Reject:       settlementservice.NewRejectPayment(deps),
			MarkPaid:     settlementservice.NewMarkPaid(deps),
			WriteLimiter: routes.writeLimiter,
		},
		SecurityLog: routes.securityLog,
	}
}

type countSettlementIssuer struct {
	issuer *settlementservice.Issuer
}

func (i countSettlementIssuer) IssueOnCount(
	ctx context.Context, partnershipID string, countedAt time.Time,
) (countdomain.SettlementRef, bool, error) {
	settlement, isIssued, err := i.issuer.IssueOnCount(ctx, partnershipID, countedAt)
	if err != nil {
		return countdomain.SettlementRef{}, false, fmt.Errorf("issue settlement on count: %w", err)
	}
	return countdomain.SettlementRef{ID: settlement.ID, Number: settlement.Number}, isIssued, nil
}

type countApprover struct {
	counts *countrepository.Counts
}

func (a countApprover) ApproveSources(ctx context.Context, approval settlementdomain.Approval) error {
	if err := a.counts.ApproveCountsOfSettlement(ctx, countdomain.SettlementApproval(approval)); err != nil {
		return fmt.Errorf("approve source counts: %w", err)
	}
	return nil
}

type outstandingSettler struct {
	issuer *settlementservice.Issuer
}

func (s outstandingSettler) SettleOutstanding(ctx context.Context, partnershipID string, at time.Time) error {
	if _, _, err := s.issuer.IssueOutstanding(ctx, partnershipID, at); err != nil {
		return fmt.Errorf("issue outstanding settlement: %w", err)
	}
	return nil
}
