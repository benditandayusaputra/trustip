package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	authrepository "github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	counthandler "github.com/benditandayusaputra/trustip/api/internal/modules/count/handler"
	drophandler "github.com/benditandayusaputra/trustip/api/internal/modules/drop/handler"
	needhandler "github.com/benditandayusaputra/trustip/api/internal/modules/need/handler"
	notificationhandler "github.com/benditandayusaputra/trustip/api/internal/modules/notification/handler"
	orghandler "github.com/benditandayusaputra/trustip/api/internal/modules/org/handler"
	partnershiphandler "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/handler"
	producthandler "github.com/benditandayusaputra/trustip/api/internal/modules/product/handler"
	proposalhandler "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/handler"
	settlementhandler "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/handler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
	"github.com/benditandayusaputra/trustip/api/internal/platform/sse"
)

const hoursPerDay = 24

type demoSecrets struct {
	peppers  security.SecretRing
	sealKeys security.ActiveKeyring
}

type demoUseCases struct {
	org         orghandler.Deps
	product     producthandler.Deps
	proposal    proposalhandler.Deps
	drop        drophandler.Deps
	count       counthandler.Deps
	settlement  settlementhandler.Deps
	need        needhandler.Deps
	partnership partnershiphandler.Deps
	inbox       notificationhandler.Deps
	jobs        jobRunners
}

type demoCast struct {
	owners       map[string]actor.Actor
	products     map[string]string
	categories   map[string]string
	needs        map[string]string
	partnerships map[string]string
	proposals    map[int]string
}

type demoReview struct {
	reviewer actor.Actor
	id       string
}

type demoSeeder struct {
	use    demoUseCases
	cast   demoCast
	world  demoWorld
	clock  *clock.Manual
	pool   *database.Pool
	hasher passwordHasher
	users  *authrepository.Users
	today  time.Time
	now    time.Time
	drops  []demoReview
	counts []demoReview
}

type quietFeed struct{}

func (quietFeed) Publish(string, sse.Event) {}

func newDemoSeeder(pool *database.Pool, secrets demoSecrets, now time.Time) (*demoSeeder, error) {
	world, err := loadDemoWorld()
	if err != nil {
		return nil, err
	}
	hasher, err := security.NewPasswordHasher(secrets.peppers, passwordHashing())
	if err != nil {
		return nil, fmt.Errorf("build demo password hasher: %w", err)
	}
	local := now.In(period.Jakarta())
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, period.Jakarta())
	manual := clock.NewManual(today.AddDate(0, 0, -world.SetupDaysAgo))
	return &demoSeeder{
		use: buildDemoUseCases(pool, manual, secrets.sealKeys), world: world, clock: manual, pool: pool,
		hasher: passwordHasher{hasher: hasher}, users: authrepository.NewUsers(pool), today: today, now: now,
		cast: demoCast{
			owners: map[string]actor.Actor{}, products: map[string]string{}, categories: map[string]string{},
			needs: map[string]string{}, partnerships: map[string]string{}, proposals: map[int]string{},
		},
	}, nil
}

func buildDemoUseCases(pool *database.Pool, manual *clock.Manual, sealKeys security.ActiveKeyring) demoUseCases {
	base := foundation{pool: pool, clock: manual, feed: quietFeed{}}
	trade := buildTradeComponents(base, sealKeys)
	routes := tradeRoutes{clock: manual}
	return demoUseCases{
		org:         orgDeps(buildOrgComponents(pool, manual), orgRoutes{}),
		product:     productDeps(buildProductComponents(pool, manual), productRoutes{}),
		proposal:    proposalDeps(trade.proposals, proposalRoutes{}),
		drop:        dropDeps(trade.drops, routes),
		count:       countDeps(trade.counts, routes),
		settlement:  settlementDeps(trade.settlements, routes),
		need:        needDeps(buildNeedDeps(base), directoryRoutes{clock: manual}),
		partnership: partnershipDeps(trade, routes),
		inbox:       notificationDeps(buildNotificationComponents(pool, manual), notificationRoutes{}),
		jobs:        buildJobRunners(base, trade),
	}
}

func (s *demoSeeder) dayAt(daysAgo int, hour int) time.Time {
	return s.today.AddDate(0, 0, -daysAgo).Add(time.Duration(hour) * time.Hour)
}

func (s *demoSeeder) moveTo(at time.Time) {
	if gap := at.Sub(s.clock.Now()); gap > 0 {
		s.clock.Advance(gap)
	}
}

func newDemoKey() (string, error) {
	key, err := idgen.UUIDv7{}.NewID()
	if err != nil {
		return "", fmt.Errorf("generate demo key: %w", err)
	}
	return key, nil
}

func (s *demoSeeder) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(s.pool.Querier(ctx))
}

func (s *demoSeeder) Run(ctx context.Context) error {
	if err := s.queries(ctx).ResetDemoData(ctx); err != nil {
		return fmt.Errorf("reset demo data: %w", err)
	}
	steps := []func(context.Context) error{
		s.loadCategories, s.foundOrganizations, s.createProducts, s.playHistory, s.finish,
	}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}
