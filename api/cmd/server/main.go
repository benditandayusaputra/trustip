package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	directoryhandler "github.com/benditandayusaputra/trustip/api/internal/modules/directory/handler"
	filehandler "github.com/benditandayusaputra/trustip/api/internal/modules/file/handler"
	insighthandler "github.com/benditandayusaputra/trustip/api/internal/modules/insight/handler"
	notificationhandler "github.com/benditandayusaputra/trustip/api/internal/modules/notification/handler"
	orghandler "github.com/benditandayusaputra/trustip/api/internal/modules/org/handler"
	sealhandler "github.com/benditandayusaputra/trustip/api/internal/modules/seal/handler"
	trusthandler "github.com/benditandayusaputra/trustip/api/internal/modules/trust/handler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/scheduler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const (
	requestsPerMinutePerIP      = 300
	verificationsPerMinutePerIP = 30
)

func main() {
	cfg, err := config.Load(os.Environ())
	if err != nil {
		logger.New(os.Stderr, slog.LevelInfo).Error("refusing to start", "error", err)
		os.Exit(1)
	}
	base := logger.New(os.Stdout, cfg.LogLevel)
	appLog := logger.ForStream(base, logger.App)
	if err := serve(cfg, base); err != nil {
		appLog.Error("server stopped", "error", err)
		os.Exit(1)
	}
	appLog.Info("server stopped")
}

func openDatabase(cfg config.Config, base *slog.Logger) (*database.Pool, error) {
	pool, err := database.Open(context.Background(), database.Options{
		URL:               cfg.DatabaseURL,
		Logger:            logger.ForStream(base, logger.App),
		Clock:             clock.System{},
		RefuseSchemaOwner: cfg.Environment == config.Production,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return pool, nil
}

func serve(cfg config.Config, base *slog.Logger) error {
	pool, err := openDatabase(cfg, base)
	if err != nil {
		return err
	}
	defer pool.Close()
	systemClock := clock.System{}
	components, err := buildAuthComponents(cfg, pool, systemClock)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	live := newLiveComponents(ctx.Done(), systemClock)
	foundation := foundation{pool: pool, clock: systemClock, feed: live.hub}
	trade := buildTradeComponents(foundation, cfg.SealKeys)
	server := newServer(base, serverDeps{
		ready:     pool.Ping,
		edge:      security.NewEdgeVerifier(cfg.EdgeKeys, systemClock),
		limiter:   ratelimit.New(ratelimit.Rule{Limit: requestsPerMinutePerIP, Window: time.Minute}, systemClock),
		clock:     systemClock,
		auth:      components,
		org:       buildOrgComponents(pool, systemClock),
		files:     buildFileComponents(pool, cfg, systemClock),
		inbox:     buildNotificationComponents(pool, systemClock),
		live:      live,
		trade:     trade,
		directory: buildDirectoryComponents(foundation),
		appOrigin: cfg.PublicAppOrigin,
	})
	appLog := logger.ForStream(base, logger.App)
	jobs := newScheduler(schedulerParts{
		runners: buildJobRunners(foundation, trade).loggingChainsTo(logger.ForStream(base, logger.Security)),
		clock:   systemClock, waits: scheduler.SystemWaits{}, log: appLog,
	})
	defer startJobs(ctx, jobs, stop)()
	appLog.Info("server starting", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
	if err := server.Run(ctx, cfg.HTTPAddr); err != nil {
		return fmt.Errorf("run http server: %w", err)
	}
	return nil
}

func startJobs(ctx context.Context, jobs *scheduler.Scheduler, stop context.CancelFunc) func() {
	var background sync.WaitGroup
	background.Go(func() { jobs.Run(ctx) })
	return func() {
		stop()
		background.Wait()
	}
}

type serverDeps struct {
	ready     httpserver.ReadinessCheck
	edge      httpserver.EdgeVerifier
	limiter   httpserver.RateLimiter
	clock     serverClock
	auth      authComponents
	org       orgComponents
	files     fileComponents
	inbox     notificationComponents
	live      liveComponents
	trade     tradeComponents
	directory directoryComponents
	appOrigin string
}

type serverClock interface {
	Now() time.Time
}

func newServer(base *slog.Logger, deps serverDeps) *httpserver.Server {
	securityLog := logger.ForStream(base, logger.Security)
	server := httpserver.New(httpserver.Options{
		AppLog:      logger.ForStream(base, logger.App),
		AccessLog:   logger.ForStream(base, logger.Access),
		SecurityLog: securityLog,
		IDs:         idgen.UUIDv7{},
		Clock:       deps.clock,
		Build:       httpserver.ReadBuildInfo(),
		Ready:       deps.ready,
		Edge:        deps.edge,
		Limiter:     deps.limiter,
		AppOrigin:   deps.appOrigin,
	})
	authhandler.Mount(server.API(), authDeps(deps.auth, securityLog, deps.clock))
	authenticate := httpserver.Authenticate(newAccessVerifier(deps.auth), deps.auth.cookies.AccessName())
	visit := httpserver.AuthenticateIfPresent(newAccessVerifier(deps.auth), deps.auth.cookies.AccessName())
	writeLimiter := newWriteLimiter(deps.clock)
	orghandler.Mount(server.API(), orgDeps(deps.org, orgRoutes{
		authenticate: authenticate,
		filesBaseURL: deps.files.publicBaseURL,
		writeLimiter: writeLimiter,
		securityLog:  securityLog,
	}))
	files := fileDeps(deps.files, fileRoutes{
		authenticate: authenticate, limiterClock: deps.clock, securityLog: securityLog,
	})
	filehandler.Mount(server.API(), files)
	filehandler.MountPublic(server.App(), files)
	notificationhandler.Mount(server.API(), notificationDeps(deps.inbox, notificationRoutes{
		authenticate: authenticate, writeLimiter: writeLimiter, securityLog: securityLog,
	}))
	httpserver.MountStream(server.API(), streamOptions(deps.live, streamRoutes{
		authenticate: authenticate, clock: deps.clock, securityLog: securityLog, appOrigin: deps.appOrigin,
	}))
	mountMarket(server.API(), deps, tradeRoutes{
		authenticate: authenticate, visit: visit, filesBaseURL: deps.files.publicBaseURL, writeLimiter: writeLimiter,
		securityLog: securityLog, clock: deps.clock,
	})
	return server
}

func mountMarket(api fiber.Router, deps serverDeps, routes tradeRoutes) {
	mountTrade(api, deps.trade, routes)
	mountDirectory(api, deps.directory, directoryRoutes{
		authenticate: routes.authenticate, visit: routes.visit, writeLimiter: routes.writeLimiter,
		filesBaseURL: routes.filesBaseURL, securityLog: routes.securityLog, clock: routes.clock,
	})
	insighthandler.Mount(api, insightDeps(deps.trade.insights, newDirectoryCandidates(deps.directory, deps.clock),
		routes))
	trusthandler.Mount(api, trustDeps(deps.trade.listings, routes))
	sealhandler.Mount(api, sealhandler.Deps{
		Verify: deps.trade.seals.public, Keys: deps.trade.seals.keys,
		Limiter: ratelimit.New(ratelimit.Rule{Limit: verificationsPerMinutePerIP, Window: time.Minute},
			routes.clock),
		SecurityLog: routes.securityLog,
	})
}

func publicRoutes() []httpserver.PublicRoute {
	routes := append(httpserver.PublicRoutes(), authhandler.PublicRoutes()...)
	routes = append(routes, orghandler.PublicRoutes()...)
	routes = append(routes, httpserver.StreamPublicRoutes()...)
	routes = append(routes, directoryhandler.PublicRoutes()...)
	routes = append(routes, sealhandler.PublicRoutes()...)
	return append(routes, filehandler.PublicRoutes()...)
}
