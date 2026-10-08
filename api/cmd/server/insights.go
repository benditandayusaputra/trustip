package main

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	directorydomain "github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	directoryservice "github.com/benditandayusaputra/trustip/api/internal/modules/directory/service"
	insightdomain "github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
	insighthandler "github.com/benditandayusaputra/trustip/api/internal/modules/insight/handler"
	insightservice "github.com/benditandayusaputra/trustip/api/internal/modules/insight/service"
)

const candidatesPerSearch = "20"

type insightComponents struct {
	histories  insightservice.DropHistoryReader
	dashboards insightservice.DashboardReader
}

func insightDeps(
	components insightComponents, candidates insightservice.CandidateFinder, routes tradeRoutes,
) insighthandler.Deps {
	return insighthandler.Deps{
		Authenticate: routes.authenticate,
		DropSuggest:  insightservice.NewSuggestDrops(components.histories, routes.clock),
		Insights:     insightservice.NewListInsights(components.dashboards, candidates, routes.clock),
		Dashboard:    insightservice.NewGetDashboard(components.dashboards, routes.clock),
	}
}

type directoryCandidates struct {
	stores     *directoryservice.ListStores
	consignors *directoryservice.ListConsignors
}

func newDirectoryCandidates(components directoryComponents, clock serverClock) directoryCandidates {
	return directoryCandidates{
		stores:     directoryservice.NewListStores(components.viewers, components.stores, clock),
		consignors: directoryservice.NewListConsignors(components.viewers, components.consignors, clock),
	}
}

func (d directoryCandidates) FindCandidates(
	ctx context.Context, signedIn actor.Actor,
) ([]insightdomain.Candidate, error) {
	if signedIn.Membership.OrganizationType == actor.Consignor {
		return d.findStores(ctx, signedIn)
	}
	return d.findConsignors(ctx, signedIn)
}

func bestMatchSearch() directorydomain.SearchForm {
	return directorydomain.SearchForm{Sort: string(directorydomain.BestMatch)}
}

func (d directoryCandidates) findStores(ctx context.Context, signedIn actor.Actor) ([]insightdomain.Candidate, error) {
	page, err := d.stores.Execute(ctx, signedIn, directoryservice.ListRequest[directorydomain.StoreSearchForm]{
		Search: directorydomain.StoreSearchForm{Search: bestMatchSearch()}, Limit: candidatesPerSearch,
	})
	if err != nil {
		return nil, fmt.Errorf("find matching stores: %w", err)
	}
	return candidatesOf(page.Items, storeCandidate), nil
}

func (d directoryCandidates) findConsignors(
	ctx context.Context, signedIn actor.Actor,
) ([]insightdomain.Candidate, error) {
	request := directoryservice.ListRequest[directorydomain.ConsignorSearchForm]{
		Search: directorydomain.ConsignorSearchForm{Search: bestMatchSearch()}, Limit: candidatesPerSearch,
	}
	page, err := d.consignors.Execute(ctx, signedIn, request)
	if err != nil {
		return nil, fmt.Errorf("find matching consignors: %w", err)
	}
	return candidatesOf(page.Items, consignorCandidate), nil
}

func candidatesOf[Card any](cards []Card, convert func(Card) insightdomain.Candidate) []insightdomain.Candidate {
	candidates := make([]insightdomain.Candidate, 0, len(cards))
	for _, card := range cards {
		candidates = append(candidates, convert(card))
	}
	return candidates
}

func storeCandidate(card directorydomain.StoreCard) insightdomain.Candidate {
	return insightdomain.Candidate{
		OrganizationID: card.ID, Name: card.Name, Score: card.Match.Score,
		IsPartner: card.IsPartner || card.IsSelf,
	}
}

func consignorCandidate(card directorydomain.ConsignorCard) insightdomain.Candidate {
	return insightdomain.Candidate{
		OrganizationID: card.ID, Name: card.Name, Score: card.Match.Score,
		IsPartner: card.IsPartner || card.IsSelf,
	}
}
