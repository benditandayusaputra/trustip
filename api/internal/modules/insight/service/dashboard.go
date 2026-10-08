package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
)

type DashboardReader interface {
	ReadSnapshot(ctx context.Context, query domain.SnapshotQuery) (domain.Snapshot, error)
	ReadDashboard(ctx context.Context, query domain.DashboardQuery) (domain.Dashboard, error)
}

type ListInsights struct {
	dashboards DashboardReader
	clock      Clock
}

func NewListInsights(dashboards DashboardReader, clock Clock) *ListInsights {
	return &ListInsights{dashboards: dashboards, clock: clock}
}

func (uc *ListInsights) Execute(ctx context.Context, signedIn actor.Actor) ([]domain.Insight, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return nil, fmt.Errorf("check insight reader: %w", err)
	}
	now := uc.clock.Now()
	snapshot, err := uc.dashboards.ReadSnapshot(ctx, domain.NewSnapshotQuery(signedIn.Membership.OrganizationID, now))
	if err != nil {
		return nil, fmt.Errorf("read insight snapshot: %w", err)
	}
	snapshot.Side = signedIn.Membership.OrganizationType
	return domain.Top(domain.Generate(snapshot, now)), nil
}

type GetDashboard struct {
	dashboards DashboardReader
	clock      Clock
}

func NewGetDashboard(dashboards DashboardReader, clock Clock) *GetDashboard {
	return &GetDashboard{dashboards: dashboards, clock: clock}
}

func (uc *GetDashboard) Execute(ctx context.Context, signedIn actor.Actor) (domain.Dashboard, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return domain.Dashboard{}, fmt.Errorf("check dashboard reader: %w", err)
	}
	query := domain.NewDashboardQuery(signedIn.Membership.OrganizationID, uc.clock.Now())
	dashboard, err := uc.dashboards.ReadDashboard(ctx, query)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("read dashboard: %w", err)
	}
	return dashboard, nil
}
