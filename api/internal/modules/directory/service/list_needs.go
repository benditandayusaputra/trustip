package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ListNeedBoard struct {
	viewers ViewerReader
	board   NeedBoard
	clock   Clock
}

func NewListNeedBoard(viewers ViewerReader, board NeedBoard, clock Clock) *ListNeedBoard {
	return &ListNeedBoard{viewers: viewers, board: board, clock: clock}
}

func (uc *ListNeedBoard) Execute(
	ctx context.Context, signedIn actor.Actor, request ListRequest[domain.SearchForm],
) (Page[domain.NeedCard], error) {
	if err := policy.RequireVisitorOrMember(signedIn); err != nil {
		return Page[domain.NeedCard]{}, fmt.Errorf("check board reader: %w", err)
	}
	search, err := domain.NewNeedSearch(request.Search)
	if err != nil {
		return Page[domain.NeedCard]{}, fmt.Errorf("read board search: %w", err)
	}
	browsing, err := startBrowsing(ctx, uc.viewers, signedIn, request)
	if err != nil {
		return Page[domain.NeedCard]{}, err
	}
	now := uc.clock.Now()
	query := func(listing domain.Listing) domain.NeedQuery {
		return domain.NeedQuery{Listing: listing, CategoryIDs: search.CategoryIDs, Text: search.Text, Now: now}
	}
	page, err := browseFrom(ctx, browsing, search.Search, source[domain.NeedCard]{
		count: func(ctx context.Context, listing domain.Listing) (map[int]int, error) {
			return uc.board.CountNeedsWithin(ctx, query(listing))
		},
		list: func(ctx context.Context, listing domain.Listing) ([]domain.NeedCard, error) {
			return uc.board.ListNeeds(ctx, query(listing))
		},
		score: func(card domain.NeedCard) domain.NeedCard { return card },
	})
	if err != nil {
		return Page[domain.NeedCard]{}, fmt.Errorf("list need board: %w", err)
	}
	return page, nil
}
