package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
)

type ReportStore interface {
	SaveReport(ctx context.Context, filing domain.Filing) (domain.Filed, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}

type ReportDeps struct {
	Reports ReportStore
	Clock   Clock
	IDs     IDGenerator
}

type ReportListing struct {
	deps ReportDeps
}

func NewReportListing(deps ReportDeps) *ReportListing {
	return &ReportListing{deps: deps}
}

func (uc *ReportListing) Execute(
	ctx context.Context, signedIn actor.Actor, form domain.ReportForm,
) (domain.Filed, error) {
	if err := policy.Require(signedIn, policy.ReportListings); err != nil {
		return domain.Filed{}, fmt.Errorf("check reporter: %w", err)
	}
	draft, err := domain.NewReport(form)
	if err != nil {
		return domain.Filed{}, fmt.Errorf("validate report: %w", err)
	}
	reportID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.Filed{}, fmt.Errorf("generate report id: %w", err)
	}
	filed, err := uc.deps.Reports.SaveReport(ctx, domain.Filing{
		ID: reportID, ReporterUserID: signedIn.UserID, ReporterOrgID: signedIn.Membership.OrganizationID,
		Draft: draft, ReportedAt: uc.deps.Clock.Now(),
	})
	if errors.Is(err, domain.ErrTargetNotFound) {
		return domain.Filed{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Filed{}, fmt.Errorf("save report on %s: %w", draft.Target.ID, err)
	}
	return filed, nil
}
