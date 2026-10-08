package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
)

type GetNeed struct {
	reader NeedReader
	clock  Clock
}

func NewGetNeed(reader NeedReader, clock Clock) *GetNeed {
	return &GetNeed{reader: reader, clock: clock}
}

func (uc *GetNeed) Execute(ctx context.Context, signedIn actor.Actor, needID string) (domain.Detail, error) {
	if err := policy.Require(signedIn, policy.ManageNeeds); err != nil {
		return domain.Detail{}, fmt.Errorf("check need reader: %w", err)
	}
	key := keyOf(signedIn, needID)
	need, err := uc.reader.GetNeed(ctx, key)
	if err != nil {
		return domain.Detail{}, fmt.Errorf("get need %s: %w", needID, err)
	}
	responses, err := uc.reader.ListResponses(ctx, key)
	if err != nil {
		return domain.Detail{}, fmt.Errorf("list responses of %s: %w", needID, err)
	}
	return domain.Detail{Need: need.AsOf(uc.clock.Now()), Responses: responses}, nil
}

type ListRequest struct {
	Status string
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.Need
	NextCursor string
	HasMore    bool
}

type ListNeeds struct {
	reader NeedReader
	clock  Clock
}

func NewListNeeds(reader NeedReader, clock Clock) *ListNeeds {
	return &ListNeeds{reader: reader, clock: clock}
}

func (uc *ListNeeds) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.Require(signedIn, policy.ManageNeeds); err != nil {
		return Page{}, fmt.Errorf("check need reader: %w", err)
	}
	status, err := domain.ParseStatus(request.Status)
	if err != nil {
		return Page{}, fmt.Errorf("read status: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	found, err := uc.reader.ListNeeds(ctx, domain.ListQuery{
		StoreID: signedIn.Membership.OrganizationID, Status: status, After: after, Limit: limit + 1,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list needs: %w", err)
	}
	now := uc.clock.Now()
	for index := range found {
		found[index] = found[index].AsOf(now)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items:      found[:limit],
		NextCursor: pagination.CursorOf(pagination.Position{At: last.CreatedAt, ID: last.ID}),
		HasMore:    true,
	}, nil
}

type ExpireNeeds struct {
	writer NeedWriter
}

func NewExpireNeeds(writer NeedWriter) *ExpireNeeds {
	return &ExpireNeeds{writer: writer}
}

func (uc *ExpireNeeds) Execute(ctx context.Context, now time.Time) (int, error) {
	expired, err := uc.writer.ExpireOpen(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("mark expired needs: %w", err)
	}
	return expired, nil
}
